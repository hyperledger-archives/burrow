### Changed
- The snatives functions have new signatures; string arguments are now string, not byte32.
- The Solidity interface contracts can be generated using the "burrow snatives" command, and the make snatives target is gone.

### Fixed
- TxExecutions that were exceptions (for example those that were REVERTed) will no longer have their events emitted from ExecutionEventsServer.GetEvents. They remain stored in state for the time being.
- CallTxSim and CallCodeSim now take same code path as real transactions (via CallContext)

### Added
- Upgraded to Tendermint [0.22.8](https://github.com/tendermint/tendermint/compare/v0.22.4...v0.22.8) (from 0.22.4).
- Support mempool signing for BroadcastTxAsync.
- Reload log file (e.g. for logrotate) on SIGHUP and dump capture logs on SIGUSR1 and on shutdown (e.g. for debug).
- File logger accepts {{.Timestamp}} in file names to generate a log file per run.
- Ability to set --external-address on burrow configure and burrow start
- Ability to set various command line options on burrow configure and burrow start and by BURROW_ prefixed environment variables
- Exposed Tendermint SeedMode option


### Fixed
- Release our mempool signing lock once transactions have been CheckTx'd' to massively increase throughput.

