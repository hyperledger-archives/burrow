# Eris-DB v0.16

|[![GoDoc](https://godoc.org/github.com/eris-db?status.png)](https://godoc.org/github.com/eris-ltd/eris-db) | Linux |
|---|-------|
| Master | [![Circle CI](https://circleci.com/gh/eris-ltd/eris-db/tree/master.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-db/tree/master) |
| Develop | [![Circle CI (develop)](https://circleci.com/gh/eris-ltd/eris-db/tree/develop.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-db/tree/develop) |

Eris-db is Monax' permissioned blockchain client. It executes Ethereum smart contracts on a permissioned virtual machine. Eris-db provides transaction finality and high transaction throughput on proof-of-stake Tendermint consensus engine. For smart contract development most functionality is provided by `eris chains`, exposed through [eris](https://monax.io/docs), the entry point for the Eris Platform.

## Table of Contents

- [What is Eris-db](#what-is-eris-db)
- [Installation](#installation)
- [For developers](#for-developers)
- [Usage](#usage)
- [Configuration](#configuration)
- [Contribute](#contribute)
- [License](#license)
- [Future work](#future-work)

## What is Eris-db ?

Eris-db is a permissioned blockchain node that executes smart contract code following the Ethereum specification.  Eris-db is built for a multi-chain universe with application specific optimization in mind. Eris-db as a node is constructed out of three main components; the consensus engine, the permissioned Ethereum virtual machine and the rpc gateway.  More specifically Eris-db consists of the following:

- **Consensus Engine:** transactions are ordered and finalised with the Byzantine fault-tolerant Tendermint protocol.  The Tendermint protocol provides high transaction throughput over a set of known validators and prevents the blockchain from forking.
- **Application Blockchain Interface (ABCI):** The smart contract application interfaces with the consensus engine over the ABCI. The ABCI allows for the consensus engine to remain agnostic from the smart contract application.
- **Smart Contract Application:** transactions are validated and applied to the application state in the order that the consensus engine has finalised them.  The application state consists of all accounts, the validator set and the name registry. Accounts in eris-db have permissions and either contain smart contract code or correspond to a public-private key pair. A transaction that calls on the smart contract code in a given account will activate the execution of that account’s code in a permissioned virtual machine.
- **Permissioned Ethereum Virtual Machine:** This virtual machine is built to observe the Ethereum operation code specification and additionally asserts the correct permissions have been granted. Permissioning is enforced through secure native functions and underlies all smart contract code. An arbitrary but finite amount of gas is handed out for every execution to ensure a finite execution duration - “You don’t need money to play, when you have permission to play”.
- **Application Binary Interface (ABI):** transactions need to be formulated in a binary format that can be processed by the blockchain node.  Currently tooling provides functionality to compile, deploy and link solidity smart contracts and formulate transactions to call smart contracts on the chain.  For proof-of-concept purposes we provide a eris-contracts.js library that automatically mirrors the smart contracts deployed on the chain and to develop middleware solutions against the blockchain network.  Future work on the light client will be aware of the ABI to natively translate calls on the API into signed transactions that can be broadcast on the network.
- **API Gateway:** eris-db exposes REST and JSON-RPC endpoints to interact with the blockchain network and the application state through broadcasting transactions, or querying the current state of the application. Websockets allow to subscribe to events, which is particularly valuable as the consensus engine and smart contract application can give unambiguously finalised results to transactions within one blocktime of about one second.

Eris-db has been architected with a longer term vision on security and data privacy from the outset:

- **Cryptographically Secured Consensus:** proof-of-stake Tendermint protocol achieves consensus over a known set of validators where every block is closed with cryptographic signatures from a majority of validators only.  No unknown variables come into play while reaching consensus on the network (as is the case for proof-of-work consensus). This guarantees that all actions on the network are fully cryptographically verified and traceable.
- **Remote Signing:** transactions can be signed by elliptic curve cryptographic algorithms, either ed25519/sha512 or secp256k1/sha256 are currently supported. Eris-db connects to a remote signing solution to generate key pairs and request signatures. Eris-keys is a placeholder for a reverse proxy into your secure signing solution. This has always been the case for transaction formulation and work continues to enable remote signing for the validator block signatures too.
- **Secure Signing:** Monax is a legal engineering company; we partner with expert companies to natively support secure signing solutions going forward.
- **Multi-chain Universe (Step 1 of 3):** from the start the eris platform has been conceived for orchestrating many chains, as exemplified by the command “eris chains make” or by that transactions are only valid on the intended chain. Separating state into different chains is only the first of three steps towards privacy on smart contract chains (see future work below).

## Installation

`eris-db` is intended to be used by the `eris chains` command via [eris](https://monax.io/docs). Available commands such as `make | start | stop | logs | inspect | update` are used for chain lifecycle management.

### For Developers
Dependency management for eris-db is managed with [glide](github.com/Masterminds/glide), and you can build eris-db from source by following

- [Install go](https://golang.org/doc/install)
- Ensure you have `gmp` installed (`sudo apt-get install libgmp3-dev || brew install gmp`)
- and execute following commands in a terminal:
- `go get github.com/Masterminds/glide`
- `go get -d github.com/eris-ltd/eris-db`
- `REPO=$($GOPATH/src/github.com/eris-ltd/eris-db)`
- `cd $REPO && glide install`
- `cd $REPO/cmd/eris-db && go install`


To run `eris-db`, just type `$ eris-db serve --work-dir <path to chain directory>`, where the chain directory needs to contain the configuration, genesis file, and private validator file as generated by `eris chains make`.

This will start the node using the provided folder as working dir. If the path is omitted it defaults to `~/.erisdb`.

For a Vagrant file see [eris-vagrant](https://github.com/eris-ltd/eris-vagrant) for drafts or soon this repo for [Vagrant](https://github.com/eris-ltd/eris-db/issues/514) and Packer files.

## Usage

Once the server has started, it will begin syncing up with the network. At that point you may begin using it. The preferred way is through our [javascript api](https://github.com/monax/eris-db.js), but it is possible to connect directly via HTTP or websocket.

## Configuration

A commented template config will be written as part of the `eris chains make` [process](https://monax.io/docs/getting-started) and can be edited prior to the `eris chains start` [process](https://monax.io/docs/getting-started).

## Contribute

We welcome all contributions and have submitted the code base to the Hyperledger project governance during incubation phase.  As an integral part of this effort we want to invite new contributors, not just to maintain but also to steer the future direction of the code in an active and open process.

You can find us on:
- [the Marmot Den (slack)](http://slack.monax.io)
- [here on Github](http://github.com/eris-ltd/eris-db/issues)
- [support.monax.io](http://support.monax.io)
- read the [Contributor file](.github/CONTRIBUTING.md)

## Future work

Some burrows marmots have already started digging to build the enterprise ecosystem applications of the future are listed in [docs/proposals](./docs/PROPOSALS.md).  Marmots live in groups we welcome your help on these or other improvement proposals. Please help us by joining the conversation on what features matter to you.

## License

[Apache 2.0](license.md)
