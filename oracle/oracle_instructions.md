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

### Manual Build (Advanced Users)

If you want to manually build the Oracle, follow the [build instructions](https://github.com/RSquad/ton-teleport-btc-periphery/blob/master/oracle/README.md) in the official repository.
**⚠️ Important:** It is strongly recommended to use MyTonCtrl for Oracle installation and management instead of manual building.

## Configuration

The oracle uses environment variables loaded from /usr/src/ton-teleport-btc-periphery/out/.env file.

### Common Variables

- `COMMON_TON_CONFIG` - The URL or local path to the TON configuration JSON file. By default, the mainnet configuration will be used ([https://ton.org/global-config.json](https://ton.org/global-config.json)).
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

## Oracle Post-Upgrade Verification Checklist

After upgrading the Oracle application, the following verification steps must be completed by validators hosting the Oracle application:

1. **Verify Oracle is started and functioning correctly.** Check the logs (refer to the .env file for log location) and ensure there are no errors. If errors are detected, please consult with the development team.

2. **Verify environment variables.** Ensure the following variables have the correct values:

    ```bash
    COMMON_TON_CONTRACT_COORDINATOR=Ef_q19o4m94xfF-yhYB85Qe6rTHDX-VTSzxBh4XpAfZMaOvk
    ORACLE_STANDALONE_MODE=false
    ```

    The variable `COMMON_TON_CONFIG` must contain a path or URL to the TON mainnet global configuration JSON file.

3. **Confirm Oracle participation in DKG generation.** Your validator must be included in the list of TON masterchain validators. If you have just started, you must wait for the next election and subsequent DKG generation (which occurs after the election).

A public API is available for monitoring DKG status:

**For mainnet:** [https://teleport.tg/metrics/api?source=dkg_status](https://teleport.tg/metrics/api?source=dkg_status)

The API returns JSON data. Focus on the `DkgInfo` section:

- **State** - Current DKG state: `FINISHED`, `IN_PROGRESS`, `PART1_FINISHED`, `PART2_FINISHED`
- **VSetSize** - TON VSet size
- **ValidatorsCountMax** - Maximum count of validators in DKG
- **ValidatorsCountInDkg** - Count of validators participating in DKG (while DKG is not FINISHED, this represents validators who have sent round1 packages)
- **ValidatorsCountNotInDkg** - Count of validators NOT participating in DKG (while DKG is not FINISHED, this represents validators who have not sent round1 packages. In FINISHED state, this list is always empty)
- **ValidatorsCountEvicted** - Count of validators evicted from the current DKG. The most common cause is validators not sending round1/round2 packages within the DKG timeout period.
- **ValidatorsIdxInDkg** - List of validators participating in DKG. Each element contains: index in VSet and ADNL address
- **ValidatorsIdxNotInDkg** - List of validators NOT participating in DKG
- **ValidatorsIdxEvicted** - List of validators evicted from the current DKG

**Verification requirement:** Wait for **State = FINISHED** and confirm that your validators are included in the **ValidatorsIdxInDkg** list.

## Troubleshooting

First, ensure that logging is enabled and set to debug mode:

1. Open the .env file at `/usr/src/ton-teleport-btc-periphery/out/.env`
2. Adjust `LOG_LEVEL` to `DEBUG` for detailed troubleshooting
3. Set `LOG_FILE` to `/var/log/oracle.txt`

```.env
...
LOG_LEVEL=DEBUG
LOG_FILE=/var/log/oracle.txt
...
```

Then restart the Oracle using `sudo systemctl restart btc_teleport` if the Oracle was installed through MyTonCtrl. Review the logs for detailed information.

## Security Considerations

- Protect the `.env` file containing sensitive keys
- Ensure proper file permissions on keystore directory
- Monitor Oracle logs for unusual activity
- Keep Oracle updated to the latest version
- Backup keystore data securely

## Support and Resources

- [Official Documentation](https://docs.ton.org/v3/guidelines/nodes/running-nodes/full-node#install-the-mytonctrl)
- [Oracle Source Code](https://github.com/RSquad/ton-teleport-btc-periphery/tree/master)
