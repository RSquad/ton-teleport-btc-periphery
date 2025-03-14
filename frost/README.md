# Go-FROST Library

A Go wrapper for the FROST (Flexible Round-Optimized Schnorr Threshold) signature scheme implemented in Rust. This library enables distributed key generation and threshold signing for Bitcoin transactions in the TON-Teleport bridge.

## Overview

The Go-FROST library provides bindings to a Rust implementation of the FROST threshold signature scheme. It supports:

- Distributed Key Generation (DKG)
- Threshold signing with tweaks (for Bitcoin taproot signatures)
- Signature verification
- Public key extraction

## Building the Library

### Prerequisites

- Go 1.23.1 or higher
- Rust and Cargo (latest stable version)
- cbindgen (for generating C headers from Rust)

### Building the Rust Component

The FROST implementation is written in Rust and compiled as a static library that is linked with the Go code:

1. Install Rust and Cargo if you haven't already:
   ```bash
   curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
   ```

2. Install cbindgen:
   ```bash
   cargo install cbindgen
   ```

3. Build the Rust library:
   ```bash
   cd rust
   ./build.sh
   ```

   This will:
   - Compile the Rust code into a static library
   - Generate C headers (`frost.h`) using cbindgen

### Building the Go Wrapper

Once the Rust library is built, you can build the Go wrapper:

```bash
go build .
```

## Testing

Run the tests to verify the implementation:

```bash
go test -v
```

