### Changed
- [JS] Significant refactor/rewrite of Burrow.js into idiomatic Typescript including some breaking changes to API
- [JS] Change to use ethers.js for ABI encoding

### Fixed
- [State] Fixed cache-concurrency bug (https://github.com/hyperledger/burrow/commit/314357e0789b0ec7033a2a419b816d2f1025cad0) and ensured consistency snapshot is used when performing simulated call reads
- [Web3] Omit empty values from JSONRPC calls

### Added
- [Tendermint] Added support for passing node options to Tendermint - e.g. custom reactors (thanks @nmanchovski!)
- [JS] Historic events can now be requested via API
- [JS] Contract deployments will now include ABIs via contract metadata so Burrow's ABI registry can be used

