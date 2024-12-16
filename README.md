# ton-teleport-btc-periphery

This repository is a periphery to the TON Teleport BTC system, containing both an Indexer and a Relayer.

## Prerequisites

### Dependencies

Before you begin, ensure you have the following installed:

- **Go:** [Install Go](https://golang.org/dl/)
- **Docker:** [Install Docker](https://docs.docker.com/get-docker/)

### Configuration

Create a `.env` file in the root directory of the project and add the following variables:

#### Common Variables

These variables are required for both the Indexer and Relayer components:

- `COMMON_BITCOIN_RPC_HOST` — The host address for the Bitcoin RPC server. Example:
  ```bash
  COMMON_BITCOIN_RPC_HOST=127.0.0.1
  ```
- `COMMON_BITCOIN_RPC_USER` — The username for authenticating with the Bitcoin RPC server. Example:
  ```bash
  COMMON_BITCOIN_RPC_USER=bitcoinrpc
  ```
- `COMMON_BITCOIN_RPC_PASS` — The password for authenticating with the Bitcoin RPC server. Example:
  ```bash
  COMMON_BITCOIN_RPC_PASS=your_rpc_password
  ```
- `COMMON_TON_CONFIG_URL` — The URL for the TON configuration. Example:
  ```bash
  COMMON_TON_CONFIG_URL=https://ton.org/config.json
  ```
- `COMMON_TON_CENTER_API_KEY` — Your API key for the TON Center API. Example:
  ```bash
  COMMON_TON_CENTER_API_KEY=your_api_key_here
  ```
- `COMMON_TON_CENTER_V3_HOST` — The host address for the TON Center V3 API. Example:
  ```bash
  COMMON_TON_CENTER_V3_HOST=https://toncenter.com/api/v3
  ```
- `COMMON_TON_CONTRACT_TELEPORT_ADDR` — The address of the Teleport contract in TON. Example:
  ```bash
  COMMON_TON_CONTRACT_TELEPORT_ADDR=EQDTpT0hDs7x8UtB49DU4eXseKEU4BzA6qUrc1cWmfy71bTR
  ```
- `COMMON_TON_CONTRACT_COORDINATOR` — The address of the Coordinator contract in TON. Example:
  ```bash
  COMMON_TON_CONTRACT_COORDINATOR=EQBrNI_Ms98m1JVq5NkC-oFRK2hPBPU27kDBC-9K_AhK759L
  ```
- `COMMON_TON_CONTRACT_BITCLIENT_ADDR` — The address of the Bitcoin Client contract in TON. Example:
  ```bash
  COMMON_TON_CONTRACT_BITCLIENT_ADDR=EQDTpT0hDs7x8UtB49DU4eXseKEU4BzA6qUrc1cWmfy71bTR
  ```

#### Indexer-Specific Variables

These variables are specific to the Indexer component:

- `INDEXER_DATABASE_URL` — The URL for the Indexer's database. Example:
  ```bash
  INDEXER_DATABASE_URL=postgres://user:password@localhost:5432/indexerdb
  ```

#### Relayer-Specific Variables

These variables are specific to the Relayer component:

- `RELAYER_WALLET_V4_SECRET` — The secret key for the Relayer's Wallet V4. Example:
  ```bash
  RELAYER_WALLET_V4_SECRET=your_wallet_v4_secret
  ```

**Note:** Ensure that all sensitive information, such as passwords and secret keys, is stored securely and not exposed in public repositories.

## Installation

1. **Clone the Repository**

   ```bash
   git clone git@github.com:RSquad/ton-teleport-btc-periphery.git
   cd ton-teleport-btc-periphery
   ```

2. **Build Lib**
   ```bash
   ./lib/gen-ton-center-v3-client.sh
   ```

## Getting Started

### Indexer

Follow these steps to set up and run the Indexer service:

#### Installation

- Follow the instructions in the **Configuration** section above.
- Follow the instructions in the **Dependencies** section above.

#### Start Indexer

```bash
go generate ./indexer
go run ./indexer/cmd/main.go
```

_Don't forget to set the environment variables in .env file._

#### Build Indexer

```bash
docker build -t ton-teleport-btc-indexer -f ./indexer/Dockerfile .
```

_Don't forget to set the environment variables when running the container._

#### Start Relayer

```bash
go run ./relayer/cmd/main.go
```

_Don't forget to set the environment variables in .env file._

#### Build Relayer

```bash
docker build -t ton-teleport-btc-relayer -f ./relayer/Dockerfile .
```

_Don't forget to set the environment variables when running the container._
