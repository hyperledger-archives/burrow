### Fixed
- [EVM] state/Cache no longer allows SetStorage on accounts that do not exist
- [GRPC] GetAccount on unknown account no longer causes a panic

### Added
- [Docker] Added solc 0.4.25 binary to docker container so that burrow deploy has what it needs to function
- [Execution] panics from executors are captured and pushed to error sink of TxExecution

