This is a major (pre-1.0.0) release that brings upgrades, safety improvements, cloud configuration, and GRPC endpoints to Burrow.

#### Breaking changes
In addition to breaking changes associated with Tendermint (see their changelog):
- State checkpointing logic has changed which has we load based on blockchain
- Event format has changed over rpc/V0 see execution/events/ package
- On-disk keys format has change from monax-keys to be more standard burrow keys
- Address format has been changed (by Tendermint and we have followed suite) - conversion is possible but simpler to regenerated keys

#### Features
- Tendermint 0.20.0
- Implemented EVM opcodes: REVERT, INVALID, SHL, SAR, SHR, RETURNDATACOPY, RETURNDATASIZE
- Add config templating with burrow configure --config-template-in --config-out
- Add config templates for kubernetes
- Integrate monax-keys as internal (default) or standalone keys service, key gen exposed over CLI
- Use GRPC for keys
- Add GRPC service for Transactor and Events
- Store ExecutionEvent by height and index in merkle tree state
- Add historical query for all time with GetEvents
- Add streaming GRPC service for ExecutionEvents with query language over tags
- Add metadata to ExecutionEvents
- Add BlockExplorer CLI for forensics
- Expose reason for REVERT
- Add last_block_info healthcheck endpoint to rpc/TM
- 
#### Improvements
- Implement checkpointing when saving application and blockchain state in commit - interrupted commit rolls burrow back to last block whereon it can catch up using Tendermint
- Maintain separate read-only tree in state so that long-running RPC request cannot block writes
- Improve state safety
- Improved input account server-side-signing
- Increase subscription reap time on rpc/V0 to 20 seconds
- Reorganise CLI
- Improve internal serialisation
- Refactor and modularise execution logic

#### Bug fixes
- Fix address generation from bytes mismatch


