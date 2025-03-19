use core::slice;
use frost_secp256k1_tr::{
    aggregate_with_tweak as frost_aggregate_with_tweak,
    keys::{
        dkg::{
            part1 as frost_dkg_part1, part2 as frost_dkg_part2, part3 as frost_dkg_part3,
            round1::{Package as Round1Package, SecretPackage as Round1SecretPackage},
            round2::{Package as Round2Package, SecretPackage as Round2SecretPackage},
        },
        KeyPackage, PublicKeyPackage, Tweak,
    },
    round1::{commit as frost_round1_commit, SigningCommitments, SigningNonces},
    round2::{sign_with_tweak as frost_round2_sign_with_tweak, SignatureShare},
    Error, Identifier, Signature, SigningPackage,
};
use rand::thread_rng;
use std::{collections::BTreeMap, ffi::c_void};

#[inline]
fn to_void<T: Sized>(obj: T) -> *const c_void {
    Box::into_raw(Box::new(obj)) as *const c_void
}

#[inline]
fn from_void<T: Sized>(ptr: *const c_void) -> Box<T> {
    unsafe { Box::from_raw(ptr as *mut T) }
}

#[repr(C)]
pub struct Buffer {
    data: *const u8,
    len: usize,
}

impl Buffer {
    fn to_slice<'a>(self) -> &'a [u8] {
        if self.data.is_null() {
            return &[];
        } else {
            let slice = unsafe { slice::from_raw_parts(self.data, self.len) };
            slice
        }
    }
}

trait FromBytes<T> {
    fn from_raw_parts(buf: *const u8, len: usize) -> Result<T, Error>;
    fn from_buf(buf: Buffer) -> Result<T, Error>;
    fn make_map(ptr: *const Pkg, len: usize) -> BTreeMap<Identifier, T>;
}

macro_rules! from_bytes_for {
    ($T: ident) => {
        impl FromBytes<$T> for $T {
            fn from_raw_parts(buf: *const u8, len: usize) -> Result<$T, Error> {
                $T::deserialize(unsafe { slice::from_raw_parts(buf, len) })
            }
            fn from_buf(buf: Buffer) -> Result<$T, Error> {
                $T::from_raw_parts(buf.data, buf.len)
            }
            fn make_map(ptr: *const Pkg, len: usize) -> BTreeMap<Identifier, $T> {
                let mut map: BTreeMap<Identifier, $T> = BTreeMap::new();
                unsafe {
                    let packages = slice::from_raw_parts(ptr, len);
                    for p in packages {
                        let identifier = Identifier::deserialize(&p.identifier).expect("Cannot deserialize Identifier");
                        let pkg = $T::from_raw_parts(p.buf, p.len).expect("Cannot deserialize Package");
                        map.insert(identifier, pkg);
                    }
                }
                map
            }
        }
    };
}

from_bytes_for!(SigningNonces);
from_bytes_for!(Round1Package);
from_bytes_for!(Round2Package);
from_bytes_for!(SignatureShare);
from_bytes_for!(SigningCommitments);
from_bytes_for!(KeyPackage);
from_bytes_for!(PublicKeyPackage);
from_bytes_for!(Signature);

#[repr(C)]
pub struct Pkg {
    identifier: [u8; 32],
    buf: *const u8,
    len: usize,
}

#[no_mangle]
pub extern "C" fn ext_get_identifier(key: u16, identifier: *mut [u8; 32]) -> i32 {
    let participant_identifier: Identifier = key.try_into().expect("should be nonzero");
    let vector = participant_identifier.serialize();

    let arr = unsafe { identifier.as_mut().unwrap() };
    arr.copy_from_slice(vector.as_slice());

    0
}

#[no_mangle]
pub extern "C" fn dkg_part1(
    identifier: &[u8; 32],
    min_signers: u16,
    max_signers: u16,
    package: *mut u8,
    package_len: i32,
    secret_package: *mut *const c_void,
) -> i32 {
    let mut rng = thread_rng();
    let ident = Identifier::deserialize(identifier).unwrap();
    match frost_dkg_part1(ident, max_signers, min_signers, &mut rng) {
        Err(err) => {
            println!("[FROST] error: {}", err);
            return -1;
        }
        Ok((s, p)) => match p.serialize() {
            Err(err) => {
                println!("[FROST] error: {}", err);
                return -2;
            }
            Ok(pkg_vec) => {
                if !package.is_null() && (package_len >= pkg_vec.len() as i32) {
                    unsafe {
                        package.copy_from(pkg_vec.as_slice().as_ptr(), pkg_vec.len());
                        *secret_package = Box::into_raw(Box::new(s)) as *const c_void;
                    }
                }
                pkg_vec.len() as i32
            }
        },
    }
}

#[no_mangle]
pub extern "C" fn dkg_part2(
    r1_secret: *const c_void,
    r1_pkgs_ptr: *const Pkg,
    r1_pkgs_len: usize,
    r2_pkgs_ptr: *mut *const Pkg,
    r2_secret: *mut *const c_void,
) -> i32 {
    if r1_secret.is_null() || r1_pkgs_ptr.is_null() || r2_pkgs_ptr.is_null() || r2_secret.is_null()
    {
        return -1;
    }
    let sp = from_void(r1_secret);
    let map = Round1Package::make_map(r1_pkgs_ptr, r1_pkgs_len);
    match frost_dkg_part2(*sp, &map) {
        Err(err) => {
            println!("[FROST] error: {}", err);
            return -2;
        }
        Ok((s, r2_map)) => {
            let mut r2_vec = Vec::with_capacity(map.len());
            for (id, pkg) in r2_map {
                let mut identifier = [0u8; 32];
                identifier.copy_from_slice(&id.serialize());
                let mut p = pkg.serialize().unwrap();
                p.shrink_to(p.len());
                r2_vec.push(Pkg {
                    identifier,
                    len: p.len(),
                    buf: p.leak().as_ptr() as *const u8,
                });
            }
            let count = r2_vec.len() as i32;
            unsafe {
                *r2_secret = to_void(s);
                *r2_pkgs_ptr = r2_vec.leak().as_ptr();
            }
            return count;
        }
    }
}

#[no_mangle]
pub extern "C" fn dkg_part3(
    r2_secret: *const c_void,
    r1_pkgs_ptr: *const Pkg,
    r1_pkgs_len: usize,
    r2_pkgs_ptr: *const Pkg,
    r2_pkgs_len: usize,
    public_key_pkg_ptr: *mut *const u8,
    public_key_pkg_len: *mut usize,
    secret_key_pkg_ptr: *mut *const u8,
    secret_key_pkg_len: *mut usize,
) -> i32 {
    if public_key_pkg_ptr.is_null()
        || secret_key_pkg_ptr.is_null()
        || r2_secret.is_null()
        || r1_pkgs_ptr.is_null()
        || r2_pkgs_ptr.is_null()
    {
        return -1;
    }
    match frost_dkg_part3(
        &from_void(r2_secret),
        &Round1Package::make_map(r1_pkgs_ptr, r1_pkgs_len),
        &Round2Package::make_map(r2_pkgs_ptr, r2_pkgs_len),
    ) {
        Err(err) => {
            println!("[FROST] error: {}", err);
            return -2;
        }
        Ok((s, p)) => {
            let mut public_vec = p.serialize().unwrap();
            public_vec.shrink_to(public_vec.len());
            let mut secret_vec = s.serialize().unwrap();
            secret_vec.shrink_to(secret_vec.len());
            unsafe {
                *public_key_pkg_len = public_vec.len();
                *public_key_pkg_ptr = public_vec.leak().as_ptr();
                *secret_key_pkg_len = secret_vec.len();
                *secret_key_pkg_ptr = secret_vec.leak().as_ptr();
            }
            return 0;
        }
    }
}

#[no_mangle]
pub extern "C" fn free_r1_secret(r1_secret: *mut c_void) {
    let _ = unsafe { Box::from_raw(r1_secret as *mut Round1SecretPackage) };
}

#[no_mangle]
pub extern "C" fn free_r2_secret(r2_secret: *mut c_void) {
    let _ = unsafe { Box::from_raw(r2_secret as *mut Round2SecretPackage) };
}

#[no_mangle]
pub extern "C" fn free_r2_pkg_vec(ptr: *const Pkg, len: usize) {
    unsafe {
        let pkgs = Vec::from_raw_parts(ptr as *mut Pkg, len, len);
        for p in pkgs {
            let _ = Vec::from_raw_parts(p.buf as *mut u8, p.len, p.len);
        }
    }
}

#[no_mangle]
pub extern "C" fn free_package_ptr(ptr: *const u8, len: usize) {
    let _ = unsafe { Vec::from_raw_parts(ptr as *mut u8, len, len) };
}

#[no_mangle]
pub extern "C" fn commit(
    key_package_buf: Buffer,
    nonces_buf: *mut Buffer,
    commitments_buf: *mut Buffer,
) -> i32 {
    match KeyPackage::from_buf(key_package_buf) {
        Err(err) => {
            println!("[FROST] KeyPackage::deserialize error: {}", err);
            return -1;
        }
        Ok(key_package) => {
            let mut rng = thread_rng();
            let (nonces, commitments) = frost_round1_commit(key_package.signing_share(), &mut rng);
            let mut nonce_vec = nonces.serialize().unwrap();
            nonce_vec.shrink_to(nonce_vec.len());
            let mut commitments_vec = commitments.serialize().unwrap();
            commitments_vec.shrink_to(commitments_vec.len());

            unsafe {
                (*nonces_buf).len = nonce_vec.len();
                (*nonces_buf).data = nonce_vec.leak().as_ptr();

                (*commitments_buf).len = commitments_vec.len();
                (*commitments_buf).data = commitments_vec.leak().as_ptr();
            }
        }
    }

    0
}

#[no_mangle]
pub extern "C" fn sign_with_tweak(
    key_package_buf: Buffer,
    message_buf: Buffer,
    commitments_ptr: *const Pkg,
    commitments_len: usize,
    nonces_buf: Buffer,
    merkle_root_buf: Buffer,
    signature_share_buf: *mut Buffer,
) -> i32 {
    match KeyPackage::from_buf(key_package_buf) {
        Err(err) => {
            println!("[FROST] KeyPackage::deserialize error: {}", err);
            return -1;
        }
        Ok(key_package) => {
            match frost_round2_sign_with_tweak(
                &SigningPackage::new(
                    SigningCommitments::make_map(commitments_ptr, commitments_len),
                    message_buf.to_slice(),
                ),
                &SigningNonces::from_buf(nonces_buf).unwrap(),
                &key_package,
                Some(merkle_root_buf.to_slice()),
            ) {
                Err(err) => {
                    println!("[FROST] error: {}", err);
                    return -2;
                }
                Ok(signature_share) => {
                    let mut signature_share_vec = signature_share.serialize();
                    signature_share_vec.shrink_to(signature_share_vec.len());
                    unsafe {
                        (*signature_share_buf).len = signature_share_vec.len();
                        (*signature_share_buf).data = signature_share_vec.leak().as_ptr();
                    }
                }
            }
        }
    }

    0
}

#[no_mangle]
pub extern "C" fn aggregate_with_tweak(
    message_buf: Buffer,
    commitments_ptr: *const Pkg,
    commitments_len: usize,
    signature_shares_ptr: *const Pkg,
    signature_shares_len: usize,
    pubkey_package_buf: Buffer,
    merkle_root_buf: Buffer,
    signature_buf: *mut Buffer,
) -> i32 {
    let signing_package = SigningPackage::new(
        SigningCommitments::make_map(commitments_ptr, commitments_len),
        message_buf.to_slice(),
    );

    let signature_shares_map =
        SignatureShare::make_map(signature_shares_ptr, signature_shares_len);

    let aggregate_with_tweak_result = frost_aggregate_with_tweak(
        &signing_package,
        &signature_shares_map,
        &PublicKeyPackage::from_buf(pubkey_package_buf).unwrap(),
        Some(merkle_root_buf.to_slice()),
    );

    match aggregate_with_tweak_result {
        Err(err) => {
            println!("[FROST] error: {}", err);
            return -2;
        }
        Ok(signature) => {
            let mut signature_vec = signature.serialize().unwrap();
            signature_vec.shrink_to(signature_vec.len());
            unsafe {
                (*signature_buf).len = signature_vec.len();
                (*signature_buf).data = signature_vec.leak().as_ptr();
            }
        }
    }

    0
}

#[no_mangle]
pub extern "C" fn verify(
    message_buf: Buffer,
    pubkey_package_buf: Buffer,
    signature_buf: Buffer,
    merkle_root_buf: Buffer,
) -> i32 {
    match Signature::from_buf(signature_buf) {
        Err(err) => {
            println!("[FROST] error: {}", err);
            -1
        }
        Ok(signature) => {
            let pubkey = PublicKeyPackage::from_buf(pubkey_package_buf).unwrap();
            let pubkey = pubkey.tweak(Some(merkle_root_buf.to_slice()));
            match pubkey
                .verifying_key()
                .verify(message_buf.to_slice(), &signature)
            {
                Err(err) => {
                    println!("[FROST] error: {}", err);
                    -2
                }
                Ok(()) => 0,
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn extract_public_key_from_package(
    pubkey_package_buf: Buffer,
    public_key: *mut Buffer,
) -> i32 {
    let pubkey_package = match PublicKeyPackage::from_buf(pubkey_package_buf) {
        Ok(x) => x,
        Err(_) => return -1,
    };
    
    let key_vec = match pubkey_package.verifying_key().serialize() {
        Ok(x) => x,
        Err(_) => return -2,
    };
    
    unsafe {
        (*public_key).len = key_vec.len();
        (*public_key).data = key_vec.leak().as_ptr();
    }
    0
}