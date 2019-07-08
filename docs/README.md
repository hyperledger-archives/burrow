# Hyperledger Burrow Documentation

Hyperledger Burrow is a permissioned Ethereum smart-contract blockchain node. It executes Ethereum EVM smart contract code (usually written in [Solidity](https://solidity.readthedocs.io)) on a permissioned virtual machine. Burrow provides transaction finality and high transaction throughput on a proof-of-stake [Tendermint](https://tendermint.com) consensus engine.

![burrow logo](assets/images/burrow.png)

1. [Installation](INSTALL.md)
2. [Logging](LOGGING.md)
3. [Quickstart](quickstart)
   * [Single full node](quickstart/single-full-node.md) - start your first chain
   * [Send transactions](quickstart/send-transactions.md) - how to communicate with your Burrow chain
   * [Deploy contracts](quickstart/deploy-contracts.md) - interact with the Ethereum Virtual Machine
   * [Multiple validators](quickstart/multiple-validators.md) - advanced consensus setup
   * [Bonding validators](quickstart/bonding-validators.md) - bonding yourself on
   * [Seed nodes](quickstart/seed-nodes.md) - add new nodes dynamically
   * [Dump / restore](design/dump-restore.md) - create a new chain with previous state
4. [Genesis](design/genesis.md)
5. [Permissions](design/permissions.md)
6. [Architecture](architecture)
   * [State](arch/state.md)
7. [Kubernetes](https://github.com/helm/charts/tree/master/stable/burrow) - bootstraps a burrow network on a Kubernetes cluster
