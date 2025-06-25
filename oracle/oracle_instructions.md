# TON Teleport BTC (TTBTC) Oracle Installation Guide

## Overview

**TON Teleport BTC (TTBTC)** is a trustless bridge between Bitcoin and TON blockchain. This guide covers the installation and configuration of the Oracle component, which is responsible for generating distributed keys and signing transactions.

The Oracle is a critical part of the TTBTC system that must run on the same machine as mainnet validators. It integrates seamlessly with [MyTonCtrl](https://github.com/ton-blockchain/mytonctrl) for easy management.

## Prerequisites

- MyTonCtrl utility installed ([Installation Guide](https://docs.ton.org/v3/guidelines/nodes/running-nodes/full-node#install-the-mytonctrl))

## Installation

### Automatic Installation

The Oracle is automatically installed when:
- Installing a new TON Validator instance with MyTonCtrl
- Upgrading an existing validator with validator mode enabled

During the upgrade process, MyTonCtrl will:
1. Download Oracle sources from the [official repository](https://github.com/RSquad/ton-teleport-btc-periphery/tree/master)
2. Compile the Oracle (written in Go and Rust)
3. Start the Oracle service

### Manual Installation/Update

To install only the Oracle or update it to a specific branch with MyTonCtrl console:

```bash
upgrade --btc-teleport [branch]
```

**Installation Location:** `/usr/src/ton-teleport-btc-periphery/out/oracle`

## Configuration

The oracle uses environment variables loaded from /usr/src/ton-teleport-btc-periphery/out/.env file.

### Common Variables

- `COMMON_TON_CONFIG` - TON configuration URL or local path to TON config json
- `COMMON_TON_CONTRACT_COORDINATOR` - Address of the Coordinator contract on TON

### Oracle-specific Variables

- `ORACLE_STANDALONE_MODE` - Boolean flag to enable standalone mode (true/false)
- `ORACLE_KEYSTORE_PATH` - Path to the keystore directory. Be sure that oracle has write access to the folder.

### Standalone Mode Variables (required if ORACLE_STANDALONE_MODE=true)

- `ORACLE_PUBKEY` - Public key in hex format
- `ORACLE_SECRET` - Secret key in hex format

### Validator Mode Variables (required if ORACLE_STANDALONE_MODE=false)

- `ORACLE_VALIDATOR_ENGINE_CONSOLE_PATH` - Path to the validator engine console (including binary validator-engine-console)
- `ORACLE_SERVER_PUBLIC_KEY_PATH` - Path to the server public key
- `ORACLE_CLIENT_PRIVATE_KEY_PATH` - Path to the client private key
- `ORACLE_VALIDATOR_SERVER_ADDR` - Address of the validator server

### Periodic Events and API Timeouts

- `ORACLE_DKG_FETCH_PERIOD` - Specifies the interval, in seconds, at which the periodic event `DKG_FETCH` is triggered. If not set, the default value is 10 seconds.
- `ORACLE_SEND_START_DKG_PERIOD` - Specifies the interval, in seconds, at which the periodic event `SEND_START_DKG` is triggered. If not set, the default value is 10 seconds.
- `ORACLE_EXECUTE_SIGN_PERIOD` - Specifies the interval, in seconds, at which the periodic event `EXECUTE_SIGN` is triggered. If not set, the default value is 10 seconds.
- `API_CALL_TIMEOUT` - Defines the maximum time (in seconds) to wait for an API call to complete. If the API does not respond within this period, the request will be terminated. If not set, the default value is 30 seconds.

### Log

- `LOG_LEVEL` - Minimum log level. Possible values: DEBUG, INFO, WARN, ERROR. If not set, the default value is INFO.
- `LOG_FILE` - Path to the log file. A full path including the file name must be specified. If not set, logs will be written only to standard output (stdout).
- `LOG_FILE_MAX_SIZE` - Maximum size of each log file, in megabytes (default 100).
- `LOG_FILE_MAX_BACKUPS` - Maximum number of backup files to retain (default 50).
- `LOG_FILE_MAX_BACKUP_AGE` - Maximum age of backup files, in days (default 365).

## Management Commands

### Status Monitoring

After installation, check the Oracle status in MyTonCtrl:
- The status will display: `Version BTC Teleport: <commit-hash> (branch)`

Example:
```txt
Version BTC Teleport: bcde501 (master)
```

### Update Oracle

```bash
upgrade --btc-teleport [branch]
```

### Remove Oracle

```bash
remove_btc_teleport [--force]
```

**Note:** Removal will fail if the current node is an active masterchain validator unless the `--force` flag is used.

## Voting Operations

### View Current Proposals

```bash
print_offers_btc_teleport_list
```

This command displays the list of current proposals from the Teleport configurator.

### Vote for Proposals

```bash
vote_offer_btc_teleport <offer-hash> [offer-hash-2 offer-hash-3 ...]
```

You can vote for multiple proposals simultaneously by providing multiple offer hashes.

### Automatic Re-voting

MyTonCtrl implements automatic re-voting functionality:
- If a proposal doesn't pass on the first attempt, MyTonCtrl will automatically re-vote in subsequent rounds
- This continues until the proposal is accepted
- Only applies to proposals you have already voted for

## Troubleshooting

### Common Issues

1. **Oracle fails to start**: Check environment variables in `.env` file
2. **Permission errors**: Ensure Oracle has write access to keystore directory
3. **Connection issues**: Verify TON config and coordinator contract address
4. **Removal blocked**: Use `--force` flag if removing from active validator

### Log Analysis

Monitor Oracle logs to diagnose issues:
- Check the log file specified in `LOG_FILE`
- Adjust `LOG_LEVEL` to `DEBUG` for detailed troubleshooting
- Review log rotation settings if logs are filling up disk space

## Security Considerations

- Protect the `.env` file containing sensitive keys
- Ensure proper file permissions on keystore directory
- Monitor Oracle logs for unusual activity
- Keep Oracle updated to the latest version
- Backup keystore data securely

## Support and Resources

- [Official Documentation](https://docs.ton.org/v3/guidelines/nodes/running-nodes/full-node#install-the-mytonctrl)
- [Oracle Source Code](https://github.com/RSquad/ton-teleport-btc-periphery/tree/master)
