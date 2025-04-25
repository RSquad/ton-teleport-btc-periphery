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

- `COMMON_TON_CONFIG` - TON configuration URL or local path to TON config json
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

#### Periodic Events and API Timeouts

- `ORACLE_DKG_FETCH_PERIOD` - Specifies the interval, in seconds, at which the periodic event `DKG_FETCH` is triggered. If not set, the default value is 6 seconds.
- `ORACLE_SEND_START_DKG_PERIOD` - Specifies the interval, in seconds, at which the periodic event `SEND_START_DKG` is triggered. If not set, the default value is 10 seconds.
- `ORACLE_EXECUTE_SIGN_PERIOD` - Specifies the interval, in seconds, at which the periodic event `EXECUTE_SIGN` is triggered. If not set, the default value is 10 seconds.
- `API_CALL_TIMEOUT` - Defines the maximum time (in seconds) to wait for an API call to complete. If the API does not respond within this period, the request will be terminated.

#### Log
If not set, the default value is 10 seconds.
- `LOG_LEVEL` - Minimum log level. Possible values: DEBUG_LEVEL, INFO_LEVEL, WARN_LEVEL, ERROR_LEVEL.
- `LOG_FILE` - Path to the log file. A full path including the file name must be specified. If not set, logs will be written only to standard output (stdout).
- `LOG_FILE_MAX_SIZE` - Maximum size of each log file, in megabytes (default 100).
- `LOG_FILE_MAX_BACKUPS` - Maximum number of backup files to retain (default 1000).
- `LOG_FILE_MAX_BACKUP_AGE` - Maximum age of backup files, in days (default 365).
- `LOG_COORDINATOR_DKG` - Dump DKG data from 

### Example .env File

Example of standalone mode (ORACLE_STANDALONE_MODE=true)
```
COMMON_TON_CONFIG=https://ton-blockchain.github.io/testnet-global.config.json
COMMON_TON_CONTRACT_COORDINATOR=EQD5URgpjt00h5x4i9MFHWX1UjmuniYPMWnYVGwmZguJ0tMh
ORACLE_STANDALONE_MODE=true
ORACLE_PUBKEY=bf6837291f771de3e60ca0b007d9346f3b0369ff059de324aea50e4054d9cb43
ORACLE_SECRET=70dc95e268e8ded2f81048a1a3dc9b500955f2234ba76f82f723a60bce270bb5bf6837291f771de3e60ca0b007d9346f3b0369ff059de324aea50e4054d9cb43
ORACLE_VALIDATOR_ENGINE_CONSOLE_PATH=
ORACLE_SERVER_PUBLIC_KEY_PATH=
ORACLE_CLIENT_PRIVATE_KEY_PATH=
ORACLE_VALIDATOR_SERVER_ADDR=
ORACLE_KEYSTORE_PATH=/path/to/keystore
ORACLE_DKG_FETCH_PERIOD=6
ORACLE_SEND_START_DKG_PERIOD=10
ORACLE_EXECUTE_SIGN_PERIOD=10
API_CALL_TIMEOUT=30
LOG_LEVEL=DEBUG
LOG_FILE=/var/logs/oracle.txt
LOG_FILE_MAX_SIZE=100
LOG_FILE_MAX_BACKUPS=1000
LOG_FILE_MAX_BACKUP_AGE=365
```

Example of using a validator (ORACLE_STANDALONE_MODE=false)
```
COMMON_TON_CONFIG=https://ton-blockchain.github.io/testnet-global.config.json
COMMON_TON_CONTRACT_COORDINATOR=EQD5URgpjt00h5x4i9MFHWX1UjmuniYPMWnYVGwmZguJ0tMh
ORACLE_STANDALONE_MODE=false
ORACLE_PUBKEY=
ORACLE_SECRET=
ORACLE_VALIDATOR_ENGINE_CONSOLE_PATH=/path/to/validator-engine-console
ORACLE_SERVER_PUBLIC_KEY_PATH=/path/to/certs/server.pub
ORACLE_CLIENT_PRIVATE_KEY_PATH=/path/to/certs/client
ORACLE_VALIDATOR_SERVER_ADDR=127.0.0.1:4441
ORACLE_KEYSTORE_PATH=/path/to/keystore
ORACLE_DKG_FETCH_PERIOD=6
ORACLE_SEND_START_DKG_PERIOD=10
ORACLE_EXECUTE_SIGN_PERIOD=10
API_CALL_TIMEOUT=30
LOG_LEVEL=DEBUG
LOG_FILE=/var/logs/oracle.txt
LOG_FILE_MAX_SIZE=100
LOG_FILE_MAX_BACKUPS=1000
LOG_FILE_MAX_BACKUP_AGE=365
```

Note: Either a relative or absolute path can be used for COMMON_TON_CONFIG instead of URL
```
COMMON_TON_CONFIG=/path/to/testnet-global.config.json
```

## Keystore

The Oracle uses a file-based keystore to securely store cryptographic materials. The keystore is organized as follows:

### Structure

```
keystore/
├── secrets/     # Contains secret packages indexed by public key
├── sessions/    # Contains session keypairs packages indexed by the DKG `Until` value.
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
