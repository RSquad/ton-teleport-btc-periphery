# Metrics

The Metrics system collects data from the TON-Teleport bridge, including contract states and transaction details. It also monitors the system and triggers alerts.

## Configuration

The Metrics is configured using environment variables. You can set these in a `.env` file in the directory with executable oracle file or provide them directly when running the application.

### Environment Variables

- `COMMON_BITCOIN_RPC_HOST` - Bitcoin RPC address (required)
- `COMMON_BITCOIN_RPC_USER` - Bitcoin RPC username (optional)
- `COMMON_BITCOIN_RPC_PASS` - Bitcoin RPC password (optional)
- `COMMON_TON_CONFIG_URL` - TON configuration URL or local path to TON config json (required)
- `COMMON_TON_CONTRACT_COORDINATOR` - Address of the Coordinator contract on TON (optional)
- `METRICS_DATABASE_URL` - Database connection URL (required)
- `METRICS_DATABASE_MAX_CONN` – Maximum number of connections to the database (optional, default: 8)
- `METRICS_DATABASE_MAX_IDLE_CONN` – Maximum number of idle connections to the database (optional, default: 8)
- `METRICS_HTTP_PORT` - HTTP port for API requests
- `METRICS_ALERTS_TEST_API_ENABLE` – if `true`, then the API for alerts testing will be available at '/metrics/alerts_testing'.
- `WRITE_DB_CHAIN_SIZE` - Size of chain messages sent from fetchers to the database writer (default 5)
- `DKG_FETCH_PERIOD` - DKG fetch period (default 10 seconds)
- `BITCOIN_CLIENT_CONTRACT_FETCH_PERIOD` - Bitcoin Client Contract fetch period (default 60 seconds)
- `BITCOIN_NETWORK_FETCH_PERIOD` - Bitcoin Network fetch period (default 59 seconds)
- `TELEPORT_CONTRACT_FETCH_PERIOD` - Teleport Contract fetch period (default 27 seconds)
- `COORDINATOR_CONTRACT_FETCH_PERIOD` - Coordinator Contract fetch period (default 12 seconds)
- `CONTRACT_BALANCES_FETCH_PERIOD` - Contract balances fetch period (default 150 seconds)
- `ALERTS_CHECK_PERIOD` - Alerts check period (default 10 seconds)
