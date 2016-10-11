# Eris DB

|[![GoDoc](https://godoc.org/github.com/eris-db?status.png)](https://godoc.org/github.com/eris-ltd/eris-db) | Linux |
|---|-------|
| Master | [![Circle CI](https://circleci.com/gh/eris-ltd/eris-db/tree/master.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-db/tree/master) |
| Develop | [![Circle CI (develop)](https://circleci.com/gh/eris-ltd/eris-db/tree/develop.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-db/tree/develop) |

Eris DB is Eris' blockchain client. It includes a permissions layer, an implementation of the Ethereum Virtual Machine, and uses Tendermint Consensus. Most functionality is provided by `eris chains`, exposed through [eris-cli](https://monax.io/docs/documentation/cli), the entry point for the Eris Platform.

## Table of Contents

- [Background](#background)
- [Installation](#installation)
- [Usage](#usage)
  - [Security](#security)
- [Contribute](#contribute)
- [License](#license)

## Background

See the [eris-db documentation](https://monax.io/docs/documentation/db/) for more information.

## Installation

`eris-db` is intended to be used by the `eris chains` command via [eris-cli](https://monax.io/docs/documentation/cli/latest/eris_chains). Available commands such as `make | start | stop | logs | inspect | update` are used for chain lifecycle management.

### For Developers

1. [Install go](https://golang.org/doc/install)
2. Ensure you have `gmp` installed (`sudo apt-get install libgmp3-dev || brew install gmp`)
3. `go get github.com/eris-ltd/eris-db/cmd/eris-db`


To run `eris-db`, just type `$ eris-db serve --work-dir <path to chain directory>`

This will start the node using the provided folder as working dir. If the path is omitted it defaults to `~/.erisdb`


## Usage

Once the server has started, it will begin syncing up with the network. At that point you may begin using it. The preferred way is through our [javascript api](https://monax.io/docs/documentation/db.js/), but it is possible to connect directly via HTTP or websocket. The JSON-RPC and web-api reference can be found [here](https://monax.io/docs/documentation/db/latest/specifications/api/).

## Configuration

See commented template config at [server_config.toml](server_config.toml). This will be written as part of the `eris chains make` [process](https://monax.io/docs/documentation/cli/latest/eris_chains_make/) and can be edited prior to the `eris chains start` [process](https://monax.io/docs/documentation/cli/latest/eris_chains_start/).

## Contribute

See the [eris platform contributing file here](https://github.com/eris-ltd/coding/blob/master/github/CONTRIBUTING.md).

## License

[GPL-3](license.md)
