# Indexer

The Indexer and tracks events of the TON-Teleport bridge, providing data to other components about pegins and pegouts, and collects metrics.

## Configuration

The Indexer is configured using environment variables. You can set these in a `.env` file in the directory with executable oracle file or provide them directly when running the application.

### Environment Variables

#### External services
- `COMMON_BITCOIN_RPC_HOST` - Bitcoin RPC address (required)
- `COMMON_BITCOIN_RPC_USER` - Bitcoin RPC username (optional)
- `COMMON_BITCOIN_RPC_PASS` - Bitcoin RPC password (optional)
- `COMMON_TON_CONFIG_URL` - TON configuration URL or local path to TON config json (required)
- `COMMON_TON_CONTRACT_TELEPORT_ADDR` - Address of the Teleport contract on TON (optional)
- `COMMON_TON_CONTRACT_COORDINATOR` - Address of the Coordinator contract on TON (optional)
- `COMMON_TON_CONTRACT_BITCLIENT_ADDR` - Address of the Bitcoin Client contract on TON (optional)
- `COMMON_TON_CONTRACT_MINTER_ADDR` - Address of the Minter contract on TON (optional)
- `INDEXER_DATABASE_URL` - Database connection URL (required)

#### Indexer features
- `RUN_MINT_SERVICE` - Flag to enable the Mint service (optional: true or false, default true)
- `RUN_EVENT_SERVICE` - Flag to enable the Event service (optional: true or false, default true)
- `RUN_PEGOUT_MANAGER` - Flag to enable the Pegout manager service (optional: true or false, default true)
- `RUN_HTTP_SERVICE` - Flag to enable the HTTP service (optional: true or false, default true)
- `RUN_METRICS_SERVICE` - Flag to enable the Metrics service (optional: true or false, default true)
- `RUN_METRICS_FETCHER_DKG` - Flag to enable the DKG Fetcher component in the Metrics service (optional: true or false, default true)
- `RUN_METRICS_FETCHER_CONTRACT_BALANCES` - Flag to enable the Contracts Balances Fetcher component in the Metrics service (optional: true or false, default true)
- `RUN_METRICS_FETCHER_CONTRACT_BITCOIN_CLIENT` - Flag to enable the Bitcoin Client Contract Fetcher component in the Metrics service (optional: true or false, default true)
- `RUN_METRICS_FETCHER_CONTRACT_TELEPORT` - Flag to enable the Teleport Contract Fetcher component in the Metrics service (optional: true or false, default true)
- `RUN_METRICS_FETCHER_CONTRACT_COORDINATOR` - Flag to enable the Coordinator Contract Fetcher component in the Metrics service (optional: true or false, default true)
- `RUN_METRICS_FETCHER_PEGOUTS` - Flag to enable the Pegouts Fetcher component in the Metrics service (optional: true or false, default true)
- `RUN_METRICS_FETCHER_BITCOIN_NETWORK` - Flag to enable the Bitcoin Network Fetcher component in the Metrics service (optional: true or false, default true)
- `METRICS_WRITE_DB_CHAIN_SIZE` - Size of chain messages sent from fetchers to the database writer (optional, default 5)
- `METRICS_DKG_FETCH_PERIOD` - DKG fetch period (optional, default 10)
- `METRICS_BITCOIN_CLIENT_CONTRACT_FETCH_PERIOD` - Bitcoin Client Contract fetch period (optional, default 60)
- `METRICS_BITCOIN_NETWORK_FETCH_PERIOD` - Bitcoin Network fetch period (optional, default 59)
- `METRICS_TELEPORT_CONTRACT_FETCH_PERIOD` - Teleport Contract fetch period (optional, default 27)
- `METRICS_COORDINATOR_CONTRACT_FETCH_PERIOD` - Coordinator Contract fetch period (optional, default 59)