### Changed
- [Tendermint] Upgraded to Tendermint 0.34.3
- [Docker] Image will now start testnet by default

### Added
- [Vent] Added support for building Vent SQL tables from Ethereum web3 JSONRPC chains (useful for oracles/state channels with layer 1)
- [Vent] Added Status to healthcheck endpoint on Vent
- [Natives] Implemented ecrecover using btcec (revised key handling)
- [Engine] Implement cross-engine dispatch
- [WASM] Implement cross-engine calls and calls to precompiles
- [WASM] Significantly extend eWASM support and implement functions
- [WASM] Add printing debug functions
- [WASM] Implement CREATE, GETTXGASPRICE, GETBLOCKDIFFICULTY, SELFDESTRUCT eWASM functions (thanks Yoongbok Lee!)
- [WASM/JS] JS library supports deploying WASM code
- [Deploy] Can specify WASM in playbook
- [EVM] Implement CHAINID and DIFFICULTY opcodes
- [Query] PEG query grammar now supports Not ("NOT") and NotEqual ("!=") operators

### Fixed
- [Deploy] Fix flaky parallel tests
- [EVM] Use correct opcode for create2 (thanks Vitali Grabovski!)
- [ABI] Check length of input before decoding (thanks Tri-stone!)
- [WASM] Constructor argument handling
- [RLP] Incorrect use of offsets for longer bytes strings
- [RLP] Use minimal encoding for length prefixes (no leading zeros)
- [Web3] Generate correct encoding hash for RawTx (ChainID in hash digest but not payload)
- [Web3] Generate canonical weird Ethereum hex
- [State] Fix read concurrency in RWTree (on which state is based) removing need for CallSim lock workaround

### Security
- Updated elliptic JS dep to 6.5.3
- Updated lodash to 4.17.19

