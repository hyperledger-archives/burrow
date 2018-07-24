This is a major (pre-1.0.0) release that introduces the ability to change the validator set through GovTx, transaction execution history, and fuller GRPC endpoint.

#### Breaking changes
- Address format has been changed (by Tendermint and we have followed suite) - conversion is possible but simpler to regenerated keys
- JSON-RPC interface has been removed
- burrow-client has been removed
- rpc/TM methods for events and broadcast have been removed

#### Features
- Tendermint 0.24.4
- GovTx GRPC service. The validator set can be now be changed.
- Enhanced GRPC services: NameReg, Transaction index, blocks service
- Events GRPC service
- Transaction Service can set value transferred

#### Improvements
- The output of "burrow keys export" can be templated

#### Bug fixes
- Fixed panic on nil bounds for blocks service


