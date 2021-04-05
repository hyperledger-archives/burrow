# EVM

Burrow's core execution engine is our Ethereum Virtual Machine (EVM) implementation. This is implemented to be compatible with the core of the EVM specification 
and so Burrow should run all Solidity code compiled with `solc`.

## MainNet Compatibility

Burrow is compatible on the smart contract and web3 RPC levels but not the network protocol, consensus, or P2P level with public Ethereum mainnet.

This affords us a great deal more flexibility and agility to support our core public permissioned side-chain use case than we would otherwise have. 
We would also argue a number of our variances are technical improvements on public Ethereum for which upgrades are much more difficult.

## EVM Version Compatibility

We aim to provide compatibility with the latest EVM version - meaning the specific 'Instruction Set Architecture'. There is not definitive delineation of the EVM
and the Ethereum network, but in practice it is usually clear where certain changes (described in [EIP documents](https://github.com/ethereum/EIPs)) are not relevant for Burrow. 
We use the following heuristic to guide the compatibility of our implementation:

- All Solidity code compiled with the latest `solc` should run on Burrow
- All opcodes defined for the EVM should be implemented in Burrow (where the opcodes assume certain consensus or network protocol fact we try to find an analogous interpretation in Burrow)

As new EIPs are released we incorporate them into Burrow. There is [current work](https://github.com/hyperledger/burrow/issues/1240) to close the gap on some of the newer 
Ethereum precompile contracts.

## Extensions

We have a notion similar to precompiled contracts that we call 'natives' whereby we mount pseudo-contracts at a particular address with functions that can be called that expose
certain native functionality of Burrow. Most prominent is access to our permissioning system. SNatives can be displayed with `burrow snatives`.

Much of the innovation that Burrow intends to offer at the smart contract level will be provided through our 'native contracts' including access to:

- WASM contracts
- Global namn/ABI registry
- Governance primitives
- Validator staking
- Time synchronisation contracts and calendars
- Oracles
- Token economic primitives

## Gas

We only use gas to bound computation; we do not extract a fee for gas used, but we will terminate execution if the gas limit passed to the EVM is exceeded. 
We expect to make the gas schedule configurable and to provide the ability to extract a fee for gas used as part of our token economic model.

## Library Usage

Burrow aims to also provide a pleasant, extensible, and liberally licensed EVM library via our `execution/evm` package. As such we try to keep the dependencies of this package minimal, 
keep the public interfaces stable, and are happy to accept changes that improve our developer ergonomics. There have already been successful integrations of Burrow with 
[Hyperledger Sawtooth](https://github.com/hyperledger/sawtooth-seth) and [Hyperledger Fabric](https://github.com/hyperledger/fabric-chaincode-evm). 
