# FROST DKG Process: Round 2 packages encryption

## Overview

In the [FROST (Flexible Round-Optimized Schnorr Threshold Signatures)](https://datatracker.ietf.org/doc/html/rfc9591) implementation, specifically in the [ZCash Foundation implementation](https://github.com/ZcashFoundation/frost), the Distributed Key Generation (DKG) process requires special security consideration during Round 2.

## Current Security Issue

In current Oracle and Coordinator implementation round2::Packages are published in the clear. If the attacker has a threshold of round2::Packages sent by a participant, they potentially can recreate that participant's secret coefficient. And with a threshold of secret coefficients they can recreate the secret being generated in the DKG.

## Encryption Solution

So we need to encrypt round2::Packages. We can generate individual keys for symmetric encryption for each pair of Oracles using Validator private key, other Validators public keys (known from config 36). 

Key generation:
1. Convert Validator key pair from Ed25519 to Curve25519 (for ECDH over Curve25519) [reference paper](https://eprint.iacr.org/2021/509.pdf).
2. Generate key for symmetric encryption with specific function (ECDH). In short: Oracle A uses self private key and public key of Oracle B. On the other side, Oracle B uses self private key and public key of Oracle A. As the result they will get the same key value.
3. Use any symmetric encryption protocol for round2::Packages.

To implement this schema we will use [libsodium](https://doc.libsodium.org/).
- For Ed25519 → Curve25519 we will use `crypto_sign_ed25519_pk_to_curve25519` and `crypto_sign_ed25519_sk_to_curve25519` explained [here](https://doc.libsodium.org/advanced/ed25519-curve25519).
- For key generation we will use scalar multiplication ([reference](https://doc.libsodium.org/advanced/scalar_multiplication)).
- For encryption we will use xchacha20poly1305 ([reference](https://doc.libsodium.org/secret-key_cryptography/aead/chacha20-poly1305/xchacha20-poly1305_construction))

Coordinator contract will require no changes. In the Oracle we will need:
1. Get access to validator private key
2. Make libsodium→GO bindings
3. In Round2 encrypt packages
4. In Round3 decrypt packages

Current status:
1. Implemented wrapper of necessary libsodium functions in C for future use in GO
2. Implemented POC in C++ of using C wrapper 
