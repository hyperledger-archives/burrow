### Changed
- [Vent] The chain id is stored in the SQL Tables
- [CLI] Command line arguments have changed

### Fixed
- [Tendermint] Disable default Tendermint TxIndexer - for which we have no use but puts extra load on DB
- [Tendermint] The CreateEmptyBlocks and CreateEmptyBlocksInterval now works
- [State] Empty blocks are not longer stored
- [State] Genesis doc is no longer persisted at every block
- [State] Store TxExecutions as single entry per block, rather than one per Event
### Add
- [Vent] vent can restore tables from vent log using new vent restore command

