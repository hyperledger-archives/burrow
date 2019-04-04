### Changed
- [Tendermint] Upgraded to 0.31.2
- [IAVL] upgraded to 0.12.2
- [Config] Tendermint.TimeoutFactor moved to Execution.TimeoutFactor (and reused for NoConsensus mode)
- [Kernel] Refactored and various exported methods changed

### Added
- [CLI] Burrow deploy can now run multiple burrow deploy files (aka playbooks) and run them in parallel
- [Consensus] Now possible to run Burrow without Tendermint in 'NoConsensus' mode by setting Tendermint.Enabled = false  for faster local testing. Execution.TimeoutFactor can be used to control how regularly Burrow commits (and is used 

### Fixed
- [Execution] Fixed uint64 underflow (when subtracting fee from balance) not protected against in CallContext
- [Tests] Various concurrency issues fixed in tests and execution tests parallelised


