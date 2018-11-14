### Fixed
- [EVM] state/Cache no longer allows SetStorage on accounts that do not exist
- [GRPC] GetAccount on unknown account no longer causes a panic

### Added
- [Execution] panics from executors are captured and pushed to error sink of TxExecution

