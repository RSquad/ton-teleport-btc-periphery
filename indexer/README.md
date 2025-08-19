# Indexer

The Indexer and tracks events of the TON-Teleport bridge, providing data to other components about pegins and pegouts.

## Configuration

The Indexer is configured using environment variables. You can set these in a `.env` file in the directory with executable oracle file or provide them directly when running the application.

### Environment Variables

- `COMMON_BITCOIN_RPC_HOST` - Bitcoin RPC address (required)
- `COMMON_BITCOIN_RPC_USER` - Bitcoin RPC username (optional)
- `COMMON_BITCOIN_RPC_PASS` - Bitcoin RPC password (optional)
- `COMMON_TON_CONFIG_URL` - TON configuration URL or local path to TON config json (required)
- `COMMON_TON_CONTRACT_COORDINATOR` - Address of the Coordinator contract on TON (optional)
- `INDEXER_DATABASE_URL` - Database connection URL (required)
- `INDEXER_DATABASE_MAX_CONN` – Maximum number of connections to the database (optional, default: 8)
- `INDEXER_DATABASE_MAX_IDLE_CONN` – Maximum number of idle connections to the database (optional, default: 8)
