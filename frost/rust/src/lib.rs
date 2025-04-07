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
use std::{collections::BTreeMap, ffi::c_void, ptr};

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

macro_rules! print_msg {
    ($msg:expr) => {
        println!("[{}:{}] {}", file!(), line!(), $msg);
    };
    ($fmt:expr, $($arg:tt)*) => {
        println!(concat!("[{}:{}] ", $fmt), file!(), line!(), $($arg)*);
    };
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
    fn make_map(ptr: *const Pkg, len: usize) -> Result<BTreeMap<Identifier, T>, Error>;
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

            fn make_map(ptr: *const Pkg, len: usize) -> Result<BTreeMap<Identifier, $T>, Error> {
                let mut map: BTreeMap<Identifier, $T> = BTreeMap::new();
                unsafe {
                    let packages = slice::from_raw_parts(ptr, len);
                    for p in packages {
                        let identifier = Identifier::deserialize(&p.identifier)?;
                        let pkg = $T::from_raw_parts(p.buf, p.len)?;
                        map.insert(identifier, pkg);
                    }
                }
                return Ok(map);
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

fn culprit_to_bytes(ref culprit: Identifier, identifier_out: &mut [u8; 32]) -> i32 {
    let culprit_data = culprit.serialize();

    if culprit_data.len() != 32 {
        print_msg!("culprit_data.len() != 32");
        return -1;
    }

    unsafe {
        ptr::copy_nonoverlapping(culprit_data.as_ptr(), identifier_out.as_mut_ptr(), 32);
    }

    -3
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

    let result = (|| -> Result<i32, Error> {
        let ident = Identifier::deserialize(identifier)?;
        let (s, p) = frost_dkg_part1(ident, max_signers, min_signers, &mut rng)?;
        let pkg_vec = p.serialize()?;

        if !package.is_null() && (package_len >= pkg_vec.len() as i32) {
            unsafe {
                package.copy_from(pkg_vec.as_slice().as_ptr(), pkg_vec.len());
                *secret_package = Box::into_raw(Box::new(s)) as *const c_void;
            }
        }

        Ok(pkg_vec.len() as i32)
    })();

    match result {
        Err(err) => {
            print_msg!("FROST error: {}", err);
            return -1;
        }
        Ok(count) => {
            return count;
        }
    }
}

#[no_mangle]
pub extern "C" fn dkg_part2(
    r1_secret: *const c_void,
    r1_pkgs_ptr: *const Pkg,
    r1_pkgs_len: usize,
    r2_pkgs_ptr: *mut *const Pkg,
    r2_secret: *mut *const c_void,
    r2_culprit_idx_out: &mut [u8; 32],
) -> i32 {
    if r1_secret.is_null() || r1_pkgs_ptr.is_null() || r2_pkgs_ptr.is_null() || r2_secret.is_null()
    {
        print_msg!("Input data error");
        return -1;
    }

    enum ResVariant {
        Count(i32),
        CulpritId(Identifier),
    }

    let result = (|| -> Result<ResVariant, Error> {
        let r1_secret_box_tmp: Box<
            frost_core::keys::dkg::round1::SecretPackage<frost_secp256k1_tr::Secp256K1Sha256TR>,
        > = from_void(r1_secret);
        let r1_secret_box = Box::clone(&r1_secret_box_tmp);
        // Prevent r2_secret_box from being freed. It must be freed manually.
        Box::leak(r1_secret_box_tmp);

        let map = Round1Package::make_map(r1_pkgs_ptr, r1_pkgs_len)?;
        let (s, r2_map) = frost_dkg_part2(*r1_secret_box, &map)?;
        let mut r2_vec = Vec::with_capacity(map.len());

        for (id, pkg) in r2_map {
            match pkg.serialize() {
                Err(err) => {
                    print_msg!("FROST error: {}", err);
                    return Ok(ResVariant::CulpritId(id));
                }
                Ok(mut p) => {
                    let mut identifier = [0u8; 32];
                    identifier.copy_from_slice(&id.serialize());

                    p.shrink_to(p.len());
                    r2_vec.push(Pkg {
                        identifier,
                        len: p.len(),
                        buf: p.leak().as_ptr() as *const u8,
                    });
                }
            }
        }
        let count = r2_vec.len() as i32;
        unsafe {
            *r2_secret = to_void(s);
            *r2_pkgs_ptr = r2_vec.leak().as_ptr();
        }

        Ok(ResVariant::Count(count))
    })();

    match result {
        Err(Error::InvalidProofOfKnowledge { ref culprit }) => {
            culprit_to_bytes(*culprit, r2_culprit_idx_out)
        }
        Err(err) => {
            print_msg!("FROST error: {}", err);
            -1
        }
        Ok(variant) => match variant {
            ResVariant::Count(count) => count,
            ResVariant::CulpritId(ref culprit) => culprit_to_bytes(*culprit, r2_culprit_idx_out),
        },
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
    r3_culprit_idx_out: &mut [u8; 32],
) -> i32 {
    if public_key_pkg_ptr.is_null()
        || secret_key_pkg_ptr.is_null()
        || r2_secret.is_null()
        || r1_pkgs_ptr.is_null()
        || r2_pkgs_ptr.is_null()
    {
        print_msg!("Input data is null");
        return -1;
    }

    let r2_secret_box_tmp: Box<
        frost_core::keys::dkg::round2::SecretPackage<frost_secp256k1_tr::Secp256K1Sha256TR>,
    > = from_void(r2_secret);
    // Prevent r2_secret_box from being freed. It must be freed manually.
    let r2_secret_box = Box::clone(&r2_secret_box_tmp);
    Box::leak(r2_secret_box_tmp);

    let result = (|| -> Result<(), Error> {
        let r1_pkgs_map = Round1Package::make_map(r1_pkgs_ptr, r1_pkgs_len)?;
        let r2_pkgs_map = Round2Package::make_map(r2_pkgs_ptr, r2_pkgs_len)?;
        let (s, p) = frost_dkg_part3(&r2_secret_box, &r1_pkgs_map, &r2_pkgs_map)?;
        let mut public_vec = p.serialize()?;
        public_vec.shrink_to(public_vec.len());
        let mut secret_vec = s.serialize()?;
        secret_vec.shrink_to(secret_vec.len());
        unsafe {
            *public_key_pkg_len = public_vec.len();
            *public_key_pkg_ptr = public_vec.leak().as_ptr();
            *secret_key_pkg_len = secret_vec.len();
            *secret_key_pkg_ptr = secret_vec.leak().as_ptr();
        }

        Ok(())
    })();

    match result {
        Err(Error::InvalidSecretShare { ref culprit }) => match culprit {
            Some(culprit_val) => culprit_to_bytes(*culprit_val, r3_culprit_idx_out),
            None => {
                print_msg!("FROST error: culprit value empty");
                -1
            }
        },
        Err(err) => {
            print_msg!("FROST error: {}", err);
            -1
        }
        Ok(()) => 0,
    }
}

#[no_mangle]
pub extern "C" fn free_r1_secret(r1_secret: *mut c_void) {
    unsafe {
        if !r1_secret.is_null() {
            let _ = Box::from_raw(r1_secret as *mut Round1SecretPackage);
        }
    };
}

#[no_mangle]
pub extern "C" fn free_r2_secret(r2_secret: *mut c_void) {
    unsafe {
        if !r2_secret.is_null() {
            let _ = Box::from_raw(r2_secret as *mut Round2SecretPackage);
        }
    };
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
    let result = (|| -> Result<(), Error> {
        let key_package = KeyPackage::from_buf(key_package_buf)?;
        let mut rng = thread_rng();
        let (nonces, commitments) = frost_round1_commit(key_package.signing_share(), &mut rng);
        let mut nonce_vec = nonces.serialize()?;
        nonce_vec.shrink_to(nonce_vec.len());
        let mut commitments_vec = commitments.serialize()?;
        commitments_vec.shrink_to(commitments_vec.len());

        unsafe {
            (*nonces_buf).len = nonce_vec.len();
            (*nonces_buf).data = nonce_vec.leak().as_ptr();

            (*commitments_buf).len = commitments_vec.len();
            (*commitments_buf).data = commitments_vec.leak().as_ptr();
        }

        Ok(())
    })();

    match result {
        Err(err) => {
            print_msg!("FROST error: {}", err);
            -1
        }
        Ok(()) => 0,
    }
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
    culprit_idx_out: &mut [u8; 32],
) -> i32 {
    let result = (|| -> Result<(), Error> {
        let key_package = KeyPackage::from_buf(key_package_buf)?;
        let signing_nonces = SigningNonces::from_buf(nonces_buf)?;
        let signing_commitments = SigningCommitments::make_map(commitments_ptr, commitments_len)?;
        let signature_share = frost_round2_sign_with_tweak(
            &SigningPackage::new(signing_commitments, message_buf.to_slice()),
            &signing_nonces,
            &key_package,
            Some(merkle_root_buf.to_slice()),
        )?;

        let mut signature_share_vec = signature_share.serialize();
        signature_share_vec.shrink_to(signature_share_vec.len());
        unsafe {
            (*signature_share_buf).len = signature_share_vec.len();
            (*signature_share_buf).data = signature_share_vec.leak().as_ptr();
        }

        Ok(())
    })();

    match result {
        Err(Error::InvalidSignatureShare { ref culprit }) => {
            culprit_to_bytes(*culprit, culprit_idx_out)
        }
        Err(Error::InvalidSecretShare { ref culprit }) => match culprit {
            Some(culprit_val) => culprit_to_bytes(*culprit_val, culprit_idx_out),
            None => {
                print_msg!("FROST error: culprit value empty");
                -1
            }
        },
        Err(err) => {
            print_msg!("FROST error: {}", err);
            -1
        }
        Ok(()) => 0,
    }
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
    culprit_idx_out: &mut [u8; 32],
) -> i32 {
    let result = (|| -> Result<(), Error> {
        let signing_commitments = SigningCommitments::make_map(commitments_ptr, commitments_len)?;
        let signing_package = SigningPackage::new(signing_commitments, message_buf.to_slice());
        let signature_shares_map =
            SignatureShare::make_map(signature_shares_ptr, signature_shares_len)?;
        let public_key_package = PublicKeyPackage::from_buf(pubkey_package_buf)?;
        let signature = frost_aggregate_with_tweak(
            &signing_package,
            &signature_shares_map,
            &public_key_package,
            Some(merkle_root_buf.to_slice()),
        )?;
        let mut signature_vec = signature.serialize()?;
        signature_vec.shrink_to(signature_vec.len());
        unsafe {
            (*signature_buf).len = signature_vec.len();
            (*signature_buf).data = signature_vec.leak().as_ptr();
        }

        Ok(())
    })();

    match result {
        Err(Error::InvalidSignatureShare { ref culprit }) => {
            culprit_to_bytes(*culprit, culprit_idx_out)
        }
        Err(Error::InvalidSecretShare { ref culprit }) => match culprit {
            Some(culprit_val) => culprit_to_bytes(*culprit_val, culprit_idx_out),
            None => {
                print_msg!("FROST error: culprit value empty");
                -1
            }
        },
        Err(err) => {
            print_msg!("FROST error: {}", err);
            -1
        }
        Ok(()) => 0,
    }
}

#[no_mangle]
pub extern "C" fn verify(
    message_buf: Buffer,
    pubkey_package_buf: Buffer,
    signature_buf: Buffer,
    merkle_root_buf: Buffer,
) -> i32 {
    let result = (|| -> Result<(), Error> {
        let signature = Signature::from_buf(signature_buf)?;
        let pubkey = PublicKeyPackage::from_buf(pubkey_package_buf)?;
        pubkey
            .tweak(Some(merkle_root_buf.to_slice()))
            .verifying_key()
            .verify(message_buf.to_slice(), &signature)?;

        Ok(())
    })();

    match result {
        Err(err) => {
            print_msg!("FROST error: {}", err);
            -1
        }
        Ok(()) => 0,
    }
}

#[no_mangle]
pub extern "C" fn extract_public_key_from_package(
    pubkey_package_buf: Buffer,
    public_key: *mut Buffer,
) -> i32 {
    let pubkey_package = match PublicKeyPackage::from_buf(pubkey_package_buf) {
        Ok(x) => x,
        Err(err) => {
            print_msg!("FROST error: {}", err);
            return -1;
        }
    };

    let key_vec = match pubkey_package.verifying_key().serialize() {
        Ok(x) => x,
        Err(err) => {
            print_msg!("FROST error: {}", err);
            return -1;
        }
    };

    unsafe {
        (*public_key).len = key_vec.len();
        (*public_key).data = key_vec.leak().as_ptr();
    }

    0
}
