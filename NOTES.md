### Changed
- [EVM] Use TxHash to allow predictable sequence numbers for account creation (allows proposal mechanism to aggregate transactions and execute in a BatchTx) - [pull request](https://github.com/hyperledger/burrow/pull/969)
- [State] Introduced MutableForest and change state layout to a streaming model that amongst other things should not blow the GRPC message size for large transactions
- [Consensus] Upgraded Tendermint to v0.30.1
- [State] Upgraded IAVL to v0.12.1
- [EVM] Integration tests upgraded to Solidity 0.5.4
- [State] All state now stored in merkle tree via MutableForest
- [State] Full validator state history stored in forest
- [Vent] Updated EventSpec table specification configuration format
- [Vent] Added support for managing Postgres triggers

### Fixed
- [Transactor] Reduce TxExecution subscription overhead
- [Transactor] Remove excessive debug subscription timeout
- [State] Fixed issue with check-pointing that could cause divergent AppHash across node restarts- [pull request](https://github.com/hyperledger/burrow/pull/985)
- [EVM] Implemented BLOCKHASH opcode
- [EVM] Used correct callee STATICCALL to fix cross-contract queries
- [Consensus] Guarded against total validator power overflow (as limited by Tendermint)

### Added
- [EVM] Implemented [CREATE2 opcode](https://eips.ethereum.org/EIPS/eip-1014)
- [EVM] Implemented [EXTCODEHASH opcode](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1052.md)
- [Accounts] Add account GetStats to query endpoint
- [Config] Expose AddrBookStrict from Tendermint
- [Deploy] burrow deploy now prints events generated during transactions
- [Deploy] burrow deploy can use key names where addresses are used
- [Governance] Added threshold-based governance via Proposal mechanism which allows entities with Root permission to propose and vote on batches of transactions to be executed on a running network with no single entity being able to do so.
- [Governance] Added command line and query introspection of proposals as well as burrow deploy support
- [Vent] Merged Vent our SQL projection and mapping layer into the Burrow repository and binary via 'burrow vent'. See [Vent Readme](./vent/README.md)
- [State] Improved read-write separation with RWTree and ImmutableForest data structures
- [State] Implemented dump/restore to port state between different version of Burrow or to compress the execution of a chain (with a proof) onto a fresh chain


