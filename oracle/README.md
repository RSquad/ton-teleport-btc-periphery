# Oracle

The Oracle component is responsible for participating in the Distributed Key Generation (DKG) process and signing Bitcoin transactions for peg-out operations in the TON-Teleport bridge.

## Overview

The Oracle uses the FROST (Flexible Round-Optimized Schnorr Threshold) signature scheme to participate in a threshold signing process. It communicates with the TON blockchain to monitor and respond to DKG and signing requests from the Coordinator contract.

## Build Requirements

- Go 1.23.1 or higher

### Prerequisites

- [Go-FROST library](../frost/README.md) - Must be built first (see FROST README for instructions)
- Access to TON blockchain (via lite client)
- Access to validator-engine-console (if not running in standalone mode)

### Building the Oracle

To build the Oracle, run:

```bash
# From the project root
go build -o out/oracle ./oracle/cmd/main.go

# Or use the provided script
./build-oracle.sh
```

## Configuration

The Oracle is configured using environment variables. You can set these in a `.env` file in the directory with executable oracle file or provide them directly when running the application.

### Required Environment Variables

#### Common Variables

- `COMMON_TON_CONFIG_URL` - TON configuration URL
- `COMMON_TON_CONTRACT_COORDINATOR` - Address of the Coordinator contract on TON

#### Oracle-specific Variables

- `ORACLE_STANDALONE_MODE` - Boolean flag to enable standalone mode (true/false)
- `ORACLE_KEYSTORE_PATH` - Path to the keystore directory. Be sure that oracle has write access to the folder.

#### Standalone Mode Variables (required if ORACLE_STANDALONE_MODE=true)

- `ORACLE_PUBKEY` - Public key in hex format
- `ORACLE_SECRET` - Secret key in hex format

#### Validator Mode Variables (required if ORACLE_STANDALONE_MODE=false)

- `ORACLE_VALIDATOR_ENGINE_CONSOLE_PATH` - Path to the validator engine console (including binary validator-engine-console)
- `ORACLE_SERVER_PUBLIC_KEY_PATH` - Path to the server public key
- `ORACLE_CLIENT_PRIVATE_KEY_PATH` - Path to the client private key
- `ORACLE_VALIDATOR_SERVER_ADDR` - Address of the validator server

### Example .env File

Example for ORACLE_STANDALONE_MODE=false
```
COMMON_TON_CONFIG_URL=https://ton-blockchain.github.io/testnet-global.config.json
COMMON_TON_CONTRACT_COORDINATOR=EQD5URgpjt00h5x4i9MFHWX1UjmuniYPMWnYVGwmZguJ0tMh
ORACLE_STANDALONE_MODE=false
ORACLE_PUBKEY=
ORACLE_SECRET=
ORACLE_KEYSTORE_PATH=/path/to/keystore
ORACLE_VALIDATOR_ENGINE_CONSOLE_PATH=/path/to/validator-engine-console
ORACLE_SERVER_PUBLIC_KEY_PATH=/path/to/certs/server.pub
ORACLE_CLIENT_PRIVATE_KEY_PATH=/path/to/certs/client
ORACLE_VALIDATOR_SERVER_ADDR=127.0.0.1:4441
```

## Keystore

The Oracle uses a file-based keystore to securely store cryptographic materials. The keystore is organized as follows:

### Structure

```
keystore/
├── secrets/     # Contains secret packages indexed by public key
└── temp/        # Contains temporary data like nonces, commitments, and signing shares
```

### Security Considerations

- The keystore directory should have restricted permissions (0700)
- Secret files are stored with restricted permissions (0600)

## Running the Oracle

```bash
# Set environment variables or use .env file

# Run the oracle
./out/oracle
```
