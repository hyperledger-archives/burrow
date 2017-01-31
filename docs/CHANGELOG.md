# Eris-DB changelog
## 0.16.0
This is a consolidation release that fixes various bugs and improves elements
of the architecture across the Eris Platform to support a quicker release 
cadence.

- Features and improvements
  - Logging system overhauled based on the central logging interface of go-kit log. Configuration lacking in this release but should be in 0.16.1. Allows powerful routing, filtering, and output options for better operations and increasing the observability of an eris blockchain. More to follow.
  - Genesis making is improved and moved into eris-db.
  - Config templating is moved into eris-db for better synchronisation of server config between the consumer of it (eris-db) and the producers of it (eris cli and other tools).
  - Some documentation updates in code and in specs.
  - Makefile added to capture conventions around building and testing and replicate them across different environments such as continuous integration systems.

- Bugfixes
  - [RPC/v0] #464 fix divergence from JSON-RPC spec
  - [CI] #366 correction to circle ci script
  - [eris-client] #378 more descriptive error message

## 0.12.0-RC3
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
