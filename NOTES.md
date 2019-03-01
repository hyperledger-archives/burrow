### Changed
- [Genesis] Use HexBytes for Genesis AppHash

### Fixed
- [Vent] Stop Vent from swallowing errors (e.g. GRPC streaming errors)
- [Consensus] Updated to patched version of Tendermint that does not pull in go-ethereum dependency
- [CLI] Removed duplicate -t flag from burrow configure


### Added
- [Kernel] Added announce message for startup and shutdown including version, key address, and other useful metadata
- [EVM] Attempt to provide REVERT reason where possible
- [Vent] --abi and --spec can be provided multiple times to provide multiple paths to search

