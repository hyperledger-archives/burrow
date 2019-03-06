### Fixed
- [State] Avoid stack traces which may be code-path-dependent or non-deterministic from being pushed to TxExecutions and so to merkle state where they can lead to breaking consensus
- [State] KVCache iterator fixed to use low, high interface as per DB, fixing CacheDB for use in Replay

### Added
- [Logging] Included height in various execution log messages
- [Transactor] Now provides SyncInfo in error message when there is a BroadcastTxSync timeout

