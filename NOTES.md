### Changed
- [CLI] 'burrow dump' renamed 'burrow dump remote'
- [Consensus] By default Burrow no longer creates empty blocks at the end of a round - though does make on every 5 minutes by default. Set CreateEmptyBlocks to "never" or omit to create no blocks unless there are transactions, or "always" to generate blocks even when there are no transactions.
- [State] Burrow state does not store empty blocks in the execution event store even when Tendermint creates them.
- [Build] 'make install_burrow' is now just 'make install'

### Fixed
- [Deploy] Always read TxExecution exception in Burrow deploy to avoid panics later on
- [Restore] Set restore transaction hash to non-zero (sha256 of original ChainID + Height)
- [Vent] --txs and --blocks now actually enable their respective tables in the Vent database
- [Consensus] Tendermint config CreateEmptyBlocks, CreateEmptyBlocksInterval now work as intended and prevent empty blocks being produced (except when needed for proof purposes) or when the interval expires (when set)

### Added
- [Dump] burrow dump now has local variant that produces a dump directly from a compatible burrow directory rather than over GRPC. If dumping/restoring between state-incompatible versions use burrow dump remote. 

