### Fixed
- [Blockchain] Persist LastBlockTime in Blockchain - before this patch LastBlockTime would only be set correctly after the first block had been received after a node is restarted - this can lead to non-determinism in the EVM via the TIMESTAMP opcode that use the LastBlockTime which is itself sourced from Tendermint's block header (from their implementation of BFT time). Implementing no empty blocks made observing this bug more likely by increasing the amount of time spent in a bad state (LastBlockTime is initially set to GenesisTime).

