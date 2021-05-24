### Changed
- [JS] Changed Burrow interface and renamed Burrow client object to to Client (merging in features needed for solts support)

### Fixed
- [JS] Fixed RLP encoding extra leading zeros on uint64 (thanks Matthieu Vachon!)
- [JS] Improved compatibility with legacy Solidity bytes types and padding conventions
- [Events] Fixed Burrow event stream wrongly switching to streaming mode for block ranges that are available in state (when the latest block is an empty block - so not stored in state)

### Added
- [JS] Added Solidity-to-Typescript code generation support (merging in solts) - this provides helpers (build.ts, api.ts) to compile Solidity files into corresponding .abi.ts files that include types for functions, events, the ABI, and EVM bytecode, and includes bindings into Burrow JS to deploy and interact with contracts via Typescript/Javascript with strong static types
- [JS] Improved interactions with events which can now be queried over any range and with strong types, see the listenerFor, reduceEvents, readEvents, and iterateEvents functions.

