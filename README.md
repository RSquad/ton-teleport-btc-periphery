# TON-Teleport BTC Periphery

This repository contains the periphery components for the TON-Teleport Bitcoin bridge, enabling secure cross-chain transactions between TON and Bitcoin networks.

## Components

### [Oracle](./oracle/README.md)
Participates in Distributed Key Generation (DKG) and threshold signing of Bitcoin transactions for peg-out operations using the FROST signature scheme. Communicates with the TON blockchain to monitor and respond to signing requests.

### [Indexer](./indexer/README.md)
Indexes and tracks events of the TON-Teleport bridge, providing data to other components about pegins and pegouts, and collects metrics

### [FROST](./frost/README.md)
A Go wrapper for the Flexible Round-Optimized Schnorr Threshold signature scheme implemented in Rust. Enables distributed key generation and threshold signing for Bitcoin transactions.

### Relayer
Monitors the Bitcoin network for newly mined blocks and relays them to the SPV client contract on the TON blockchain.

### Lib
Shared library containing common utilities and interfaces used by all components, including TON blockchain interaction, logging, and configuration management.

## Getting Started

Each component has its own build requirements and configuration. Please refer to the individual README files linked above for detailed instructions.

## Prerequisites

- Go 1.23.1 or higher
- Rust and Cargo (for FROST)
- Access to TON and Bitcoin networks

