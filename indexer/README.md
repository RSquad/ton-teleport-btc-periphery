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
- `INDEXER_DATABASE_URL` - Database connection URL (required)
- `INDEXER_DATABASE_MAX_CONN` – Maximum number of connections to the database (optional, default: 8)
> **Note:** Two separate connection pools are created — one for GraphQL and another for all other services.  
> This means the total number of active connections can be up to `INDEXER_DATABASE_MAX_CONN * 2`.
- `INDEXER_DATABASE_MAX_IDLE_CONN` – Maximum number of idle connections to the database (optional, default: 8)
- `COMMON_TON_CONTRACT_TELEPORT_ADDR` - Address of the Teleport contract on TON (optional)
- `COMMON_TON_CONTRACT_COORDINATOR` - Address of the Coordinator contract on TON (optional)
- `COMMON_TON_CONTRACT_BITCLIENT_ADDR` - Address of the Bitcoin Client contract on TON (optional)
- `COMMON_TON_CONTRACT_MINTER_ADDR` - Address of the Minter contract on TON (optional)

#### Internal services
- `RUN_SERVICES` - Comma-separated list of service names with their run flags. By default, all services are started. You can specify which services to run by setting them to `true` or `false`.

  **Available services:**
  - `MINT_SERVICE` - Mint service
  - `EVENT_SERVICE` - Event service
  - `PEGOUT_MANAGER` - Pegout service
  - `HTTP_SERVICE` - HTTP service
  - `METRICS_SERVICE` - Metrics service
  
  **Example:** `RUN_SERVICES="METRICS_SERVICE=false,PEGOUT_MANAGER=true"`
  
  This will start all services except the Metrics Service.

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

#### Dependencies

For normal operation of the system, all services (including metrics) are essential. By default, you only need to specify the External services configuration.
In special cases (like debugging or testing), you can run only a subset of the services. Each service has dependencies on external services and requires specific environment variables:

- `MINT_SERVICE` - depends on: TON Client, Bitcoin Client, Teleport contract. Required environment variables:
   - `COMMON_TON_CONFIG_URL`
   - `COMMON_BITCOIN_RPC_HOST`
   - `COMMON_BITCOIN_RPC_USER`
   - `COMMON_BITCOIN_RPC_PASS`
   - `COMMON_TON_CONTRACT_TELEPORT_ADDR`
   - `INDEXER_DATABASE_URL`
- `EVENT_SERVICE` - depends on: TON Client, Coordinator contract, Teleport contract. Required environment variables:
   - `COMMON_TON_CONFIG_URL`
   - `COMMON_TON_CONTRACT_COORDINATOR`
   - `COMMON_TON_CONTRACT_TELEPORT_ADDR`
   - `INDEXER_DATABASE_URL`
- `PEGOUT_MANAGER` - depends on: TON Client, Bitcoin Client, Teleport contract. Required environment variables:
   - `COMMON_TON_CONFIG_URL`
   - `COMMON_BITCOIN_RPC_HOST`
   - `COMMON_BITCOIN_RPC_USER`
   - `COMMON_BITCOIN_RPC_PASS`
   - `COMMON_TON_CONTRACT_TELEPORT_ADDR`
   - `INDEXER_DATABASE_URL`
- `HTTP_SERVICE` - depends on: TON Client, Bitcoin Client, Teleport contract. Required environment variables:
   - `COMMON_TON_CONFIG_URL`
   - `COMMON_BITCOIN_RPC_HOST`
   - `COMMON_BITCOIN_RPC_USER`
   - `COMMON_BITCOIN_RPC_PASS`
   - `COMMON_TON_CONTRACT_TELEPORT_ADDR`
   - `INDEXER_DATABASE_URL`
- `METRICS_SERVICE` - each metric has its own dependencies, plus HTTP_SERVICE must be started.
   - `DKG` - depends on: TON Client, Coordinator contract. Required environment variables:
      - `COMMON_TON_CONFIG_URL`
      - `COMMON_TON_CONTRACT_COORDINATOR`
      - `INDEXER_DATABASE_URL`
   - `CONTRACT_BALANCES` - depends on: TON Client and optionally on all contract addresses. Required environment variables:
      - `COMMON_TON_CONFIG_URL`
      - `COMMON_TON_CONTRACT_COORDINATOR`
      - `INDEXER_DATABASE_URL`
      - `COMMON_TON_CONTRACT_TELEPORT_ADDR` - optional
      - `COMMON_TON_CONTRACT_COORDINATOR` - optional
      - `COMMON_TON_CONTRACT_BITCLIENT_ADDR` - optional
      - `COMMON_TON_CONTRACT_MINTER_ADDR` - optional
   - `CONTRACT_BITCOIN_CLIENT` - depends on: TON Client, Bitcoin Client, Bitcoin Client contract. Required environment variables:
      - `COMMON_TON_CONFIG_URL`
      - `COMMON_BITCOIN_RPC_HOST`
      - `COMMON_BITCOIN_RPC_USER`
      - `COMMON_BITCOIN_RPC_PASS`
      - `COMMON_TON_CONTRACT_BITCLIENT_ADDR`
      - `INDEXER_DATABASE_URL`
   - `CONTRACT_TELEPORT` - depends on: TON Client, Teleport contract. Required environment variables:
      - `COMMON_TON_CONFIG_URL`
      - `COMMON_TON_CONTRACT_TELEPORT_ADDR`
      - `INDEXER_DATABASE_URL`
   - `CONTRACT_COORDINATOR` - depends on: TON Client, Coordinator contract. Required environment variables:
      - `COMMON_TON_CONFIG_URL`
      - `COMMON_TON_CONTRACT_COORDINATOR`
      - `INDEXER_DATABASE_URL`
   - `PEGOUTS` - depends on: TON Client, Bitcoin Client, Coordinator contract. Required environment variables:
      - `COMMON_TON_CONFIG_URL`
      - `COMMON_BITCOIN_RPC_HOST`
      - `COMMON_BITCOIN_RPC_USER`
      - `COMMON_BITCOIN_RPC_PASS`
      - `COMMON_TON_CONTRACT_COORDINATOR`
      - `INDEXER_DATABASE_URL`      
   - `BITCOIN_NETWORK` - depends on: Bitcoin Client. Required environment variables:
      - `COMMON_BITCOIN_RPC_HOST`
      - `COMMON_BITCOIN_RPC_USER`
      - `COMMON_BITCOIN_RPC_PASS`
      - `INDEXER_DATABASE_URL`      
