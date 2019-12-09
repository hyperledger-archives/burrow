### Security
- [Tendermint] Upgraded to v0.32.8, checkTxAsync now includes node ID
		
### Changed
- [Vent] Sync every block height to DB and send height notification from _vent_chain table so downstream can check DB sync without --blocks
- [RPC/Query] GetName now returns GRPC NotFound status (rather than unknown) when a requested key is not set.

### Fixed
- [Execution] Simulated calls (e.g. query contracts) now returns the height of the state on which the query was run. Useful for downstream sync.
- 

