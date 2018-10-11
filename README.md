# Hyperledger Burrow

[![version](https://img.shields.io/github/tag/hyperledger/burrow.svg)](https://github.com/hyperledger/burrow/releases/latest)
[![GoDoc](https://godoc.org/github.com/burrow?status.png)](https://godoc.org/github.com/hyperledger/burrow)
[![license](https://img.shields.io/github/license/hyperledger/burrow.svg)](LICENSE.md)
[![LoC](https://tokei.rs/b1/github/hyperledger/burrow?category=lines)](https://github.com/hyperledger/burrow)

Branch    | Linux
----------|------
| Master | [![Circle CI](https://circleci.com/gh/hyperledger/burrow/tree/master.svg?style=svg)](https://circleci.com/gh/hyperledger/burrow/tree/master) |
| Develop | [![Circle CI (develop)](https://circleci.com/gh/hyperledger/burrow/tree/develop.svg?style=svg)](https://circleci.com/gh/hyperledger/burrow/tree/develop) |

Hyperledger Burrow is a permissioned Ethereum smart-contract blockchain node. It executes Ethereum EVM smart contract code (usually written in [Solidity](https://solidity.readthedocs.io)) on a permissioned virtual machine. Burrow provides transaction finality and high transaction throughput on a proof-of-stake [Tendermint](https://tendermint.com) consensus engine.

![burrow logo](docs/assets/images/burrow.png)

## What is Burrow

Hyperledger Burrow is a permissioned blockchain node that executes smart contract code following the Ethereum specification. Burrow is built for a multi-chain universe with application specific optimization in mind. Burrow as a node is constructed out of three main components: the consensus engine, the permissioned Ethereum virtual machine and the rpc gateway. More specifically Burrow consists of the following:

- **Consensus Engine:** Transactions are ordered and finalised with the Byzantine fault-tolerant Tendermint protocol.  The Tendermint protocol provides high transaction throughput over a set of known validators and prevents the blockchain from forking.
- **Application Blockchain Interface (ABCI):** The smart contract application interfaces with the consensus engine over the [ABCI](https://github.com/tendermint/tendermint/abci). The ABCI allows for the consensus engine to remain agnostic from the smart contract application.
- **Smart Contract Application:** Transactions are validated and applied to the application state in the order that the consensus engine has finalised them. The application state consists of all accounts, the validator set and the name registry. Accounts in Burrow have permissions and either contain smart contract code or correspond to a public-private key pair. A transaction that calls on the smart contract code in a given account will activate the execution of that account’s code in a permissioned virtual machine.
- **Permissioned Ethereum Virtual Machine:** This virtual machine is built to observe the Ethereum operation code specification and additionally asserts the correct permissions have been granted. Permissioning is enforced through secure native functions and underlies all smart contract code. An arbitrary but finite amount of gas is handed out for every execution to ensure a finite execution duration - “You don’t need money to play, when you have permission to play”.
- **Application Binary Interface (ABI):** Transactions need to be formulated in a binary format that can be processed by the blockchain node. Current tooling provides functionality to compile, deploy and link solidity smart contracts and formulate transactions to call smart contracts on the chain.
- **API Gateway:** Burrow exposes REST and JSON-RPC endpoints to interact with the blockchain network and the application state through broadcasting transactions, or querying the current state of the application. Websockets allow subscribing to events, which is particularly valuable as the consensus engine and smart contract application can give unambiguously finalised results to transactions within one blocktime of about one second.

## Project Roadmap

Project information generally updated on a quarterly basis can be found on the [Hyperledger Burrow Wiki](https://wiki.hyperledger.org/projects/burrow).

## Minimum requirements

Requirement|Notes
---|---
Go version | Go1.10 or higher

## Installation

See the [install instructions](docs/INSTALL.md).

## Quick Start
1. [Single full node](docs/quickstart/single-full-node.md) - start your first chain
1. [Send transactions](docs/quickstart/send-transactions.md) - how to communicate with your Burrow chain
1. [Deploy contracts](docs/quickstart/deploy-contracts.md) - interact with the Ethereum Virtual Machine
1. [Multiple validators](docs/quickstart/multiple-validators.md) - advanced consensus setup
1. [Seed nodes](docs/quickstart/seed-nodes.md) - add new node dynamically
1. [Kubernetes](https://github.com/helm/charts/tree/master/stable/burrow) - bootstraps a burrow network on a Kubernetes cluster

## Project documentation
Burrow getting started documentation is available in the [docs](docs/README.md) directory and in [GoDocs](https://godoc.org/github.com/hyperledger/burrow).

## Contribute

We welcome any and all contributions. Read the [contributing file](.github/CONTRIBUTING.md) for more information on making your first Pull Request to Burrow!

You can find us on:
- [Hyperledger Chat](https://chat.hyperledger.org)
- [Hyperledger Mailing List](https://lists.hyperledger.org/mailman/listinfo)
- [here on Github](https://github.com/hyperledger/burrow/issues)

## Future work

For some (slightly outdated) ideas on future work, see the [proposals document](docs/PROPOSALS.md).

## License

[Apache 2.0](LICENSE.md)
