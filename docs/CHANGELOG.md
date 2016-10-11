# Eris-DB changelog

## 0.12.0-RC3
This release marks a separation of Eris-DB (the blockchain node of the
 Eris platform) from our consensus engine Tendermint. This involved 
 significant refactoring of almost all parts of the code, but provides
 a solid foundation to build against using future iterations of 
 Tendermint as well as advanced permissioned smart contract execution.
 
 Many changes are under the hood but here are the main externally 
 visible changes:

- Features and improvements
  - Upgrade to Tendermint 0.6.0 in-process consensus
  - Support DELEGATECALL opcode in Ethereum Virtual Machine (important for solidity library calls)
  - ARM support
  - Docker image size reduced
  - Introduction of eris-client companion library for interacting with
  eris:db
  - Improved configuration handling using Viper
  - Allow multiple event subscriptions from same host under rpv/tendermint
 
- Tool changes  
  - Use glide instead of godeps for dependencies
  
- Testing
  - Separation of integration for RPC and unit tests

- Bugfixes
  - [EVM] Fix calculation of child CALL gaslimit (allowing solidity library calls to work properly)
  - [RPC/v0] Fix blocking event subscription in transactAndHold (preventing return in Javascript libraries)
  - [Blockchain] Fix getBlocks to respect block height cap
