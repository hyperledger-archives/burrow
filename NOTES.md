### Changed
- [Config] Reverted rename of ValidatorAddress to Address in config (each Burrow node has a specific validator key it uses for signing whether or not it is running as a validator right now)

### Fixed 
- [EVM] Return integer overflow error code (not stack overflow) for integer overflow errors
- [Docs] Fix broken examples
- [Deploy] Set input on QueryContract jobs correctly
- [EVM] Fix byte-printing for DebugOpcodes run mode
- [Crypto] Use Tendermint-compatible secp256k1 addressing
- [Natives] Make natives first class contracts and establish Dispatcher and Callable as a common calling convention for natives, EVM, and WASM (pending for WASM).
- [Natives] Fix Ethereum precompile addresses (addresses were padded on right instead of the left)


### Added
- [Web3] Implemented Ethereum web3 JSON RPC including sendRawTransaction!
- [Docs] Much docs (see also: https://www.hyperledger.org/blog/2019/10/08/burrow-the-boring-blockchain)
- [Docs] Generate github pages docs index with docsify: https://hyperledger.github.io/burrow/
- [JS] Publish burrow.js to @hyperledger/burrow
- [State] Store EVM ABI and contract metadata on-chain see [GetMetadata](https://github.com/hyperledger/burrow/blob/e80aad5d8fac1f67dbfec61ea75670f9a38c61a1/protobuf/rpcquery.proto#L25)
- [Tendermint] Upgrade to v0.32.3
- [Execution] Added IdentifyTx for introducing nodes (binding their NodeID to ValidatorAddress)
- [Natives] Implement Ethereum precompile number 5 - modular exponentiation


