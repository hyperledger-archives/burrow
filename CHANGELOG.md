# Eris-db changelog
## v0.16.0
This is a consolidation release that fixes various bugs and improves elements
of the architecture across the Eris Platform to support a quicker release
cadence.

#### Features and improvements (among others)
- [pull-510](https://github.com/eris-ltd/eris-db/pull/510) upgrade consensus engine to Tendermint v0.8.0
- [pull-507](https://github.com/eris-ltd/eris-db/pull/507) use sha3 for snative addresses for future-proofing
- [pull-506](https://github.com/eris-ltd/eris-db/pull/506) alignment and consolidation for genesis and config between tooling and chains
- [pull-504](https://github.com/eris-ltd/eris-db/pull/504) relicense eris-db to Apache 2.0
- [pull-500](https://github.com/eris-ltd/eris-db/pull/500) introduce more strongly types secure native contracts
- [pull-499](https://github.com/eris-ltd/eris-db/pull/499) introduce word256 and remove dependency on tendermint/go-common
- [pull-493](https://github.com/eris-ltd/eris-db/pull/493) re-introduce GenesisTime in GenesisDoc

- Logging system overhauled based on the central logging interface of go-kit log. Configuration lacking in this release but should be in 0.16.1. Allows powerful routing, filtering, and output options for better operations and increasing the observability of an eris blockchain. More to follow.
- Genesis making is improved and moved into eris-db.
- Config templating is moved into eris-db for better synchronisation of server config between the consumer of it (eris-db) and the producers of it (eris cli and other tools).
- Some documentation updates in code and in specs.
- [pull-462](https://github.com/eris-ltd/eris-db/pull/499) Makefile added to capture conventions around building and testing and replicate them across different environments such as continuous integration systems.

#### Bugfixes (among others)
- [pull-516](https://github.com/eris-ltd/eris-db/pull/516) Organize and add unit tests for rpc/v0
- [pull-453](https://github.com/eris-ltd/eris-db/pull/453) Fix deserialisation for BroadcastTx on rpc/v0
- [pull-476](https://github.com/eris-ltd/eris-db/pull/476) patch EXTCODESIZE for native contracts as solc ^v0.4 performs a safety check for non-zero contract code
- [pull-468](https://github.com/eris-ltd/eris-db/pull/468) correct specifications for params on unsubscribe on rpc/tendermint
- [pull-465](https://github.com/eris-ltd/eris-db/pull/465) fix divergence from JSON-RPC spec for Response object
- [pull-366](https://github.com/eris-ltd/eris-db/pull/366) correction to circle ci script
- [pull-379](https://github.com/eris-ltd/eris-db/pull/379) more descriptive error message for eris-client

## v0.12.0
This release marks the start of Eris-DB as the full permissioned blockchain node
 of the Eris platform with the Tendermint permissioned consensus engine.
 This involved significant refactoring of almost all parts of the code,
 but provides a solid foundation to build the next generation of advanced
 permissioned smart contract blockchains.

 Many changes are under the hood but here are the main externally
 visible changes:

- Features and improvements
  - Upgrade to Tendermint 0.6.0 in-process consensus
  - Support DELEGATECALL opcode in Ethereum Virtual Machine (important for solidity library calls)
  - ARM support
  - Docker image size reduced
  - Introduction of eris-client companion library for interacting with
  eris:db
  - Improved single configuration file for all components written by eris-cm
  - Allow multiple event subscriptions from same host under rpc/tendermint


- Tool changes  
  - Use glide instead of godeps for dependencies


- Testing
  - integration tests over simulated RPC calls
  - significantly improved unit tests
  - the ethereum virtual machine and the consensus engine are now top-level
  components and are exposed to continuous integration tests


- Bugfixes (incomplete list)
  - [EVM] Fix calculation of child CALL gaslimit (allowing solidity library calls to work properly)
  - [RPC/v0] Fix blocking event subscription in transactAndHold (preventing return in Javascript libraries)
  - [Blockchain] Fix getBlocks to respect block height cap
