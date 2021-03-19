### Changed
- [Repo] main branch replaces master as per Hyperledger TSC guidelines

### Fixed
- [Docker] Make sure default testnet mode works when running docker images
- [Vent] Use appropriately sized database integral types for EVM integer types (i.e. numeric for uint64 and bigger)
- [Vent] Ethereum block consumer now correctly reads to an _inclusive_ block batch end height
- [Web3] Handle integer ChainID for web3 consistently; return hex-encoded numeric value from web3 RPC, also allow overriding of genesis-hash derived ChainID so Burrow can be connected with from metamask

### Added
- [Build] Build dev docker and JS releases by force pushing to prerelease branch
- [Vent] Expose BlockConsumerConfig to adjust backoff and read characteristics
- [Vent] Add vent-side continuity test over blocks (to double-check exactly once delivery of events)

