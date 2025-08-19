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
- `COMMON_TON_CONTRACT_COORDINATOR` - Address of the Coordinator contract on TON (optional)
- `DATABASE_URL` - Database connection URL (required)
- `DATABASE_MAX_CONN` – Maximum number of connections to the database (optional, default: 8)
- `DATABASE_MAX_IDLE_CONN` – Maximum number of idle connections to the database (optional, default: 8)

#### Metrics

- `METRICS` - Comma-separated list of metric names with their run flags. By default, all metrics are started. You can specify which metrics to run by setting them to `true` or `false`.

  **Available metrics:**
  - `DKG`
  - `CONTRACT_BALANCES`
  - `CONTRACT_BITCOIN_CLIENT`
  - `CONTRACT_TELEPORT`
  - `CONTRACT_COORDINATOR`
  - `PEGOUTS`
  - `BITCOIN_NETWORK`

  **Example:** `METRICS="DKG=false,CONTRACT_BALANCES=true"`

  This will start all metrics except the DKG.

 - `METRICS_ARGS` - Comma-separated list of metric fine-tuning arguments with thair values.

   **Available arguments:**
   - `WRITE_DB_CHAIN_SIZE` - Size of chain messages sent from fetchers to the database writer (default 5)
   - `DKG_FETCH_PERIOD` - DKG fetch period (default 10)
   - `BITCOIN_CLIENT_CONTRACT_FETCH_PERIOD` - Bitcoin Client Contract fetch period (default 60)
   - `BITCOIN_NETWORK_FETCH_PERIOD` - Bitcoin Network fetch period (default 59)
   - `TELEPORT_CONTRACT_FETCH_PERIOD` - Teleport Contract fetch period (default 27)
   - `COORDINATOR_CONTRACT_FETCH_PERIOD` - Coordinator Contract fetch period (default 59)
