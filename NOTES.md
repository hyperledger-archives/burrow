### Changed
- [Consensus] Tendermint timeouts configurable by a float factor from defaults and default change to 0.33 of Tendermint's default for smaller networks'
- [Transactor] Hard-coded timeout removed from transactor and added to TxEnvelopeParam for client specified server-side timeout (in case of longer confirmation times such as when some validators are unavailable
- [Logging] ExcludeTrace config inverted to Trace and now defaults to false (i.e. no trace/debug logging). Default log output now excludes Tendermint logging (and is therefore much less talkative)

### Added
- [Logging] Add height to all logging messages
- [RPC] Add LastBlockCommitDuration to SyncInfo

### Fixed
- [Metrics] Replace use of Summary metrics when Histogram was intended

