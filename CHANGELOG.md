# [Hyperledger Burrow](https://github.com/hyperledger/burrow) Changelog
## [0.29.3] - 2019-10-16
### Changed
- [NPM] Point package.json to index.js


## [0.29.2] - 2019-10-15
### Changed
- [NPM] Publish with index.js in TLD


## [0.29.1] - 2019-10-10
### Changed
- [State] Split metadata and account state to be kinder to downstream EVM integrators


## [0.29.0] - 2019-10-08
### Changed
- [Config] Reverted rename of ValidatorAddress to Address in config (each Burrow node has a specific validator key it uses for signing whether or not it is running as a validator right now)

### Fixed 
- [EVM] Return integer overflow error code (not stack overflow) for integer overflow errors
- [Docs] Fix broken examples
- [Deploy] Set input on QueryContract jobs correctly
- [EVM] Fix byte-printing for DebugOpcodes run mode
- [Crypto] Use Tendermint-compatible secp256k1 addressing
- [Natives] Make natives first class contracts and establish Dispatcher and Callable as a common calling convention for natives, EVM, and WASM (pending for WASM).
- [Natives] Fix Ethereum precompile addresses (addresses were padded on right instead of the left)


### Added
- [Web3] Implemented Ethereum web3 JSON RPC including sendRawTransaction!
- [Docs] Much docs (see also: https://www.hyperledger.org/blog/2019/10/08/burrow-the-boring-blockchain)
- [Docs] Generate github pages docs index with docsify: https://hyperledger.github.io/burrow/
- [JS] Publish burrow.js to @hyperledger/burrow
- [State] Store EVM ABI and contract metadata on-chain see [GetMetadata](https://github.com/hyperledger/burrow/blob/e80aad5d8fac1f67dbfec61ea75670f9a38c61a1/protobuf/rpcquery.proto#L25)
- [Tendermint] Upgrade to v0.32.3
- [Execution] Added IdentifyTx for introducing nodes (binding their NodeID to ValidatorAddress)
- [Natives] Implement Ethereum precompile number 5 - modular exponentiation



## [0.28.2] - 2019-08-21
### Fixed
- [Vent] The new decode event ABI _before_ filter provides more keys but means vent must have access to all possible LogEvent ABIs when it is started. This is not practical in general so we now will will only err if an event matches but we have no ABI. This means we might not notice we have forgot to include an ABI since an event that _would_ have matched on an ABI spec field (prefixed 'Event') will not just not match, and so fail silently.


## [0.28.1] - 2019-08-21
### Fixed
- [Vent] Log for _vent_log insert now faithfully captures what is being inserted
- [Vent] Remove arbitrary 100 character limits on system table text fields

### Added
- [JS] Burrow.js now included in Burrow repo and tested with Burrow CI! Future burrow.js releases will now match version of Burrow.


## [0.28.0] - 2019-08-14
### Changed
- [State] IterateStreamEvents now takes inclusive start and end points (end used to be exclusive) avoid bug-prone conversion
- [Dump] Improved structure and API
- [Dump] Default to JSON output and use protobuf for binary output

### Fixed
- [Dump] Fix dump missing events emitted at end height provided
- [Dump] EVM events were not dumped if no height was provided to burrow dump remote commandline
- [RPC/Info] Fix panic in /names and implement properly - now accepts a 'regex' parameter which is a regular expression to match names. Empty for all names.
- [Configure] burrow configure flags --separate-genesis-doc and --pool now work together

### Added
- [State] Burrow now remembers contact ABIs (which describe how to pack bits when calling contracts) - burrow deploy and vent will both use chain-hosted ABI if they are available
- [State] Bond and unbond transactions are now implement to allow validators to transfer native token into validator power.
- [Dump] Better tests, mock, and benchmarks - suitable for profiling IAVL
- [Events] Filters now support OR connective
- [Vent] Projection filters can now have filters longer than 100 characters.
- [Vent] Falls back to local ABI
- [CLI/RPC] Contracts now hold metadata, including contract name, source file, and function names



## [0.27.0] - 2019-06-23
### Added
- [WASM] Support for WASM contracts written in Solidity compiled using solang

### Fixed
-[RPC/Transact] CallCodeSim and CallTxSim were run against uncommitted checker state rather than committed state were all other reads are routed. They were also passed through Transactor for no particularly good reason. This changes them to run against committed DB state and removes the code path through Transactor.

### Changed
- [State] TxExecution's Envelope now stored in state so will be reproduced in Vent Tx tables and over RPC, and so matches TxExecutions served from *Sync rpctransact methods


## [0.26.2] - 2019-06-19
### Fixed
- [Blockchain] Persist LastBlockTime in Blockchain - before this patch LastBlockTime would only be set correctly after the first block had been received after a node is restarted - this can lead to non-determinism in the EVM via the TIMESTAMP opcode that use the LastBlockTime which is itself sourced from Tendermint's block header (from their implementation of BFT time). Implementing no empty blocks made observing this bug more likely by increasing the amount of time spent in a bad state (LastBlockTime is initially set to GenesisTime).


## [0.26.1] - 2019-06-16
### Changed
- [CLI] 'burrow dump' renamed 'burrow dump remote'
- [Consensus] By default Burrow no longer creates empty blocks at the end of a round - though does make on every 5 minutes by default. Set CreateEmptyBlocks to "never" or omit to create no blocks unless there are transactions, or "always" to generate blocks even when there are no transactions.
- [State] Burrow state does not store empty blocks in the execution event store even when Tendermint creates them.
- [Build] 'make install_burrow' is now just 'make install'

### Fixed
- [Deploy] Always read TxExecution exception in Burrow deploy to avoid panics later on
- [Restore] Set restore transaction hash to non-zero (sha256 of original ChainID + Height)
- [Vent] --txs and --blocks now actually enable their respective tables in the Vent database
- [Consensus] Tendermint config CreateEmptyBlocks, CreateEmptyBlocksInterval now work as intended and prevent empty blocks being produced (except when needed for proof purposes) or when the interval expires (when set)

### Added
- [Dump] burrow dump now has local variant that produces a dump directly from a compatible burrow directory rather than over GRPC. If dumping/restoring between state-incompatible versions use burrow dump remote.


## [0.26.0] - 2019-06-14
### Changed
- [Vent] The chain id is stored in the SQL Tables
- [CLI] Command line arguments have changed

### Fixed
- [Tendermint] Disable default Tendermint TxIndexer - for which we have no use but puts extra load on DB
- [Tendermint] The CreateEmptyBlocks and CreateEmptyBlocksInterval now works
- [State] Empty blocks are not longer stored
- [State] Genesis doc is no longer persisted at every block
- [State] Store TxExecutions as single entry per block, rather than one per Event
### Add
- [Vent] vent can restore tables from vent log using new vent restore command


## [0.25.1] - 2019-05-03
### Changed
- [Config] Split ListenAddress into ListenHost and ListenPort to ease parsing in the Helm charts
- [CLI] Burrow restore now always fails if state is detected but can be made --silent
- [CLI] No dump client timeout by default
- [Deploy] Reduced the default logging level to trace instead of info
- [Build] Switched to Go modules

### Fixed
- [Keys] Resolved an issue where the keyStore wasn't built when using the remote keys client.
- [Deploy] Fix nil dereference in query error path, check constructor args in BuildJob


## [0.25.0] - 2019-04-05
### Changed
- [Tendermint] Upgraded to 0.31.2
- [IAVL] upgraded to 0.12.2
- [Config] Tendermint.TimeoutFactor moved to Execution.TimeoutFactor (and reused for NoConsensus mode)
- [Kernel] Refactored and various exported methods changed

### Added
- [CLI] Introduced burrow configure --pool for generation of multiple validator configs suitable for running on a single (or many) machines
- [CLI] Burrow deploy can now run multiple burrow deploy files (aka playbooks) and run them in parallel
- [Consensus] Now possible to run Burrow without Tendermint in 'NoConsensus' mode by setting Tendermint.Enabled = false  for faster local testing. Execution.TimeoutFactor can be used to control how regularly Burrow commits (and is used

### Fixed
- [Execution] Fixed uint64 underflow (when subtracting fee from balance) not protected against in CallContext
- [Tests] Various concurrency issues fixed in tests and execution tests parallelised



## [0.24.6] - 2019-03-19
### Changed
- [RPC] 'blocks' on info RPC now lists blocks in ascending rather than descending height order

### Added
- [CLI] Introduced burrow configure --pool for generation of multiple validator configs suitable for running on a single (or many) machines

### Fixed
- [Metrics] Fix histogram statistics by making counts cumulative


## [0.24.5] - 2019-03-14
### Changed
- [Consensus] Tendermint timeouts configurable by a float factor from defaults and default change to 0.33 of Tendermint's default for smaller networks'
- [Transactor] Hard-coded timeout removed from transactor and added to TxEnvelopeParam for client specified server-side timeout (in case of longer confirmation times such as when some validators are unavailable
- [Logging] ExcludeTrace config inverted to Trace and now defaults to false (i.e. no trace/debug logging). Default log output now excludes Tendermint logging (and is therefore much less talkative)

### Added
- [Logging] Add height to all logging messages
- [RPC] Add LastBlockCommitDuration to SyncInfo

### Fixed
- [Metrics] Replace use of Summary metrics when Histogram was intended


## [0.24.4] - 2019-03-08
### Changed
- [EVM] Accept []byte nonce rather than enforcing the use of txs.Tx.TxHash()
- [Crypto] Expose SequenceNonce helper to allow library users to use sequence-number based addresses for newly created contracts


## [0.24.3] - 2019-03-06
### Fixed
- [State] Avoid stack traces which may be code-path-dependent or non-deterministic from being pushed to TxExecutions and so to merkle state where they can lead to breaking consensus
- [State] KVCache iterator fixed to use low, high interface as per DB, fixing CacheDB for use in Replay

### Added
- [Logging] Included height in various execution log messages
- [Transactor] Now provides SyncInfo in error message when there is a BroadcastTxSync timeout


## [0.24.2] - 2019-02-28
### Changed
- [Genesis] Use HexBytes for Genesis AppHash

### Fixed
- [Vent] Stop Vent from swallowing errors (e.g. GRPC streaming errors)
- [Consensus] Updated to patched version of Tendermint that does not pull in go-ethereum dependency
- [CLI] Removed duplicate -t flag from burrow configure


### Added
- [Kernel] Added announce message for startup and shutdown including version, key address, and other useful metadata
- [EVM] Attempt to provide REVERT reason where possible
- [Vent] --abi and --spec can be provided multiple times to provide multiple paths to search


## [0.24.1] - 2019-02-28
### Changed
- [ABI] abi.EncodeFunctionCall and AbiSpec.Pack now take a variadic ...interface{} type for function arguments rather than []string

### Fixed
- [Deploy] Binary files are now written atomically to prevent issue with dependency libraries being momentarily truncated when deploying in parallel

### Added
- [ABI] DecodeFunctionReturn re-exposed (formerly Packer then packer in 0.24.0) to make deploy API symmetrical


## [0.24.0] - 2019-02-26
### Changed
- [EVM] Use TxHash to allow predictable sequence numbers for account creation (allows proposal mechanism to aggregate transactions and execute in a BatchTx) - [pull request](https://github.com/hyperledger/burrow/pull/969)
- [State] Introduced MutableForest and change state layout to a streaming model that amongst other things should not blow the GRPC message size for large transactions
- [Consensus] Upgraded Tendermint to v0.30.1
- [State] Upgraded IAVL to v0.12.1
- [EVM] Integration tests upgraded to Solidity 0.5.4
- [State] All state now stored in merkle tree via MutableForest
- [State] Full validator state history stored in forest
- [Vent] Updated EventSpec table specification configuration format
- [Vent] Added support for managing Postgres triggers

### Fixed
- [Transactor] Reduce TxExecution subscription overhead
- [Transactor] Remove excessive debug subscription timeout
- [State] Fixed issue with check-pointing that could cause divergent AppHash across node restarts- [pull request](https://github.com/hyperledger/burrow/pull/985)
- [EVM] Implemented BLOCKHASH opcode
- [EVM] Used correct callee STATICCALL to fix cross-contract queries
- [Consensus] Guarded against total validator power overflow (as limited by Tendermint)

### Added
- [EVM] Implemented [CREATE2 opcode](https://eips.ethereum.org/EIPS/eip-1014)
- [EVM] Implemented [EXTCODEHASH opcode](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1052.md)
- [Accounts] Add account GetStats to query endpoint
- [Config] Expose AddrBookStrict from Tendermint
- [Deploy] burrow deploy now prints events generated during transactions
- [Deploy] burrow deploy can use key names where addresses are used
- [Governance] Added threshold-based governance via Proposal mechanism which allows entities with Root permission to propose and vote on batches of transactions to be executed on a running network with no single entity being able to do so.
- [Governance] Added command line and query introspection of proposals as well as burrow deploy support
- [Vent] Merged Vent our SQL projection and mapping layer into the Burrow repository and binary via 'burrow vent'. See [Vent Readme](./vent/vent.md)
- [State] Improved read-write separation with RWTree and ImmutableForest data structures
- [State] Implemented dump/restore to port state between different version of Burrow or to compress the execution of a chain (with a proof) onto a fresh chain



## [0.23.3] - 2018-12-19
### Fixed
- [State] Since State hash is not unique (i.e if we make no writes) by storing the CommitID by AppHash we can overwrite an older CommitID with a newer one leading us to load the wrong tree version to overwrite in case of loading from a checkpoint.


## [0.23.2] - 2018-12-18
Hotfix release for 0.23.1
### Fixed
- [State] Fixed issue with checkpointing whereby RWTree would load its readTree from one version lower than it should.



## [0.23.1] - 2018-11-14
### Fixed
- [EVM] state/Cache no longer allows SetStorage on accounts that do not exist
- [GRPC] GetAccount on unknown account no longer causes a panic

### Added
- [Docker] Added solc 0.4.25 binary to docker container so that burrow deploy has what it needs to function
- [Execution] panics from executors are captured and pushed to error sink of TxExecution


## [0.23.0] - 2018-11-09
### Changed
- [ABI] provides fast event lookup of EventID
- [Events] BlockExecution now included full Tendermint block header as protobuf object rather than JSON string
- [EVM] Nested call errors are now transmitted to EventSink (e.g. TxExecution) as events for better tracing and tests
- [SNative] Permissions contract returns permission flag set not resultant permissions from setBase unsetBase and setGlobal
- [EVM] Errors transmitted through errors.Pusher interface for more reliable capture from memory, stack, and elsewhere
- [Governance] Breaking change to state structure due to governance storage in tree (state root hashes will not match)


### Fixed
- [EVM] Issue where value was not transferred because VM call state was not synced
- [EVM] Various issue where errors were swallowed (in particular - where calling an empty account and when a TX was invalid on delivery)
- [EVM] When calling a non-existent account CreateAccount permission is checked on the caller not the caller's caller
- [CLI] Version now contains date and commit
- [Test] Burrow integration test runner shuts down Burrow correctly
- [Serialisation] updated tmthrgd/go-hex to fallback on default encoding when lacking SSE 4.1 CPU instructions


### Added
- [Deploy] Burrow deploy meta jobs reuses GRPC connection
- [Governance] Added proposal mechanism (via ProposalTx) that allows bulk atomic update of smart contracts and changing network parameters via a threshold voting mechanism. This allows some level of network evolution without any single trusted party or hard forks. This should be considered alpha level functionality.
- [EVM] Added EVM State interface removing unnecessary cache layer (fixing various issues)
- [EVM] Implemented STATICCALL opcode
- [P2P] Added AuthorizedPeers config option to sync only with whitelisted peers exposed over ABCI query under key /p2p/filter/
- [EVM] stack depth now dynamically allocated and exponentially grown in the same way as memory
- [EVM] Solidity proxy call forwarding test

### Removed
- MutableAccount and ConcreteAccount


## [0.22.0] - 2018-09-21
### Changed
- Upgraded to Tendermint 0.24.0
- Upgraded to IAVL 0.11.0

### Fixed
- Fixed non-determinism in Governance Tx
- Fixed various abi issues

### Added
- burrow deploy displays revert reason when available
- burrow deploy compiles contracts concurrently

## [0.21.0] - 2018-08-21
### Changed
- Upgraded to Tendermint 0.23.0
- Validator Set Power now takes Address
- RPC/TM config renamed to RPC/Info

### Added
- Burrow deploy creates devdoc
- Docker image has org.label-schema labels

### Fixed
- Upgrade to IAVL 0.10.0 and load previous versions immutably on boot - for chains with a long history > 20 minute load times could be observed because every previous root was being loaded from DB rather than lightweight version references as was intended
- Metrics server does not panic on empty block metas and recovers from other panics


## [0.20.1] - 2018-08-17
### Changed
- The snatives functions have new signatures; string arguments are now string, not byte32.
- The Solidity interface contracts can be generated using the "burrow snatives" command, and the make snatives target is gone.

### Fixed
- TxExecutions that were exceptions (for example those that were REVERTed) will no longer have their events emitted from ExecutionEventsServer.GetEvents. They remain stored in state for the time being.
- CallTxSim and CallCodeSim now take same code path as real transactions (via CallContext)
- Release our mempool signing lock once transactions have been CheckTx'd' to massively increase throughput.

### Added
- Upgraded to Tendermint [0.22.8](https://github.com/tendermint/tendermint/compare/v0.22.4...v0.22.8) (from 0.22.4).
- Support mempool signing for BroadcastTxAsync.
- Reload log file (e.g. for logrotate) on SIGHUP and dump capture logs on SIGUSR1 and on shutdown (e.g. for debug).
- File logger accepts {{.Timestamp}} in file names to generate a log file per run.
- Ability to set --external-address on burrow configure and burrow start
- Ability to set various command line options on burrow configure and burrow start and by BURROW_ prefixed environment variables
- Exposed Tendermint SeedMode option


## [0.20.0] - 2018-07-24
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



## [0.19.0] - 2018-06-26
This is a major (pre-1.0.0) release that brings upgrades, safety improvements, cloud configuration, and GRPC endpoints to Burrow.

#### Breaking changes
In addition to breaking changes associated with Tendermint (see their changelog):
- State checkpointing logic has changed which has we load based on blockchain
- Event format has changed over rpc/V0 see execution/events/ package
- On-disk keys format has change from monax-keys to be more standard burrow keys
- Address format has been changed (by Tendermint and we have followed suite) - conversion is possible but simpler to regenerated keys

#### Features
- Tendermint 0.21.0
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



## [0.18.1]
This is a minor release including:
- Introduce InputAccount param for RPC/v0 for integration in JS libs
- Resolve some issues with RPC/tm tests swallowing timeouts and not dealing with reordered events

## [0.18.0] - 2018-05-09
This is an extremely large release in terms of lines of code changed addressing several years of technical debt. Despite this efforts were made to maintain external interfaces as much as possible and an extended period of stabilisation has taken place on develop.

A major strand of work has been in condensing previous Monax tooling spread across multiple repos into just two. The Hyperledger Burrow repo and [Bosmarmot](http://github.com/monax/bosmarmot). Burrow is now able to generate chains (replacing 'monax chains make') with 'burrow spec' and 'burrow configure'. Our 'EPM' contract deployment and testing tool, our javascript libraries, compilers, and monax-keys are avaiable in Bosmarmot (the former in the 'bos' tool). Work is underway to pull monax-keys into the Burrow project, and we will continue to make Burrow as self-contained as possible.

#### Features
- Substantial support for latest EVM and solidity 0.4.21+ (missing some opcodes that will be added shortly - see known issues)
- Tendermint 0.18.0
- All signing through monax-keys KeyClient connection (preparation for HSM and GPG based signing daemon)
- Address-based signing (Burrow acts as delegate when you send transact, transactAndHold, send, sendAndHold, and transactNameReg a parameter including input_account (hex address) instead of priv_key.
- Provide sequential signing when using transact family methods (above) - allowing 100s Tx per second with the same input account
- Genesis making, config making, and key generation through 'burrow spec' and 'burrow configure'
- Logging configuration language and text/template for output
- Improved CLI UX and framework (mow.cli)
- Improved configuration


#### Internal Improvements
- Refactored execution and provide interfaces for executor
- Segregate EVM and blockchain state to act as better library
- Panic recovery on TX execution
- Stricter interface boundaries and immutability of core objects by default
- Replace broken BlockCache with universal StateCache that doesn't write directly to DB
- All dependencies upgraded, notably: tendermint/IAVL 0.7.0
- Use Go dep instead of glide
- PubSub event hub with query language
- Heavily optimised logging
- PPROF profiling server option
- Additional tests in multiple packages including v0 RPC and concurrency-focussed test
- Use Tendermint verifier for PrivValidator
- Use monax/relic for project history
- Run bosmarmot integration tests in CI
- Update documentation
- Numerous maintainability, naming, and aesthetic code improvements

#### Bug fixes
- Fix memory leak in BlockCache
- Fix CPU usage in BlockCache
- Fix SIGNEXTEND for negative numbers
- Fix multiple execution level panics
- Make Transactor work during tendermint recheck

#### Known issues
- Documentation rot - some effort has been made to update documentation to represent the current state but in some places it has slipped help can be found (and would be welcomed) on: [Hyperledger Burrow Chat](https://chat.hyperledger.org/channel/burrow)
- Missing support for: RETURNDATACOPY and RETURNDATASIZE https://github.com/hyperledger/burrow/issues/705 (coming very soon)
- Missing support for: INVALID https://github.com/hyperledger/burrow/issues/705 (coming very soon)
- Missing support for: REVERT https://github.com/hyperledger/burrow/issues/600 (coming very soon)


## [0.17.1]
Minor tweaks to docker build file

## [0.17.0] - 2017-09-04
This is a service release with some significant ethereum/solidity compatibility improvements and new logging features. It includes:

- [Upgrade to use Tendermint v0.9.2](https://github.com/hyperledger/burrow/pull/595)
- [Implemented dynamic memory](https://github.com/hyperledger/burrow/pull/607) assumed by the EVM bytecode produce by solidity, fixing various issues.
- Logging sinks and configuration - providing a flexible mechanism for configuring log flows and outputs see [logging section in readme](https://github.com/hyperledger/burrow#logging). Various other logging enhancements.
- Fix event unsubscription
- Remove module-specific versioning
- Rename suicide to selfdestruct
- SNative tweaks

Known issues:

- SELFDESTRUCT opcode causes a panic when an account is removed. A [fix](https://github.com/hyperledger/burrow/pull/605) was produced but was [reverted](https://github.com/hyperledger/burrow/pull/636) pending investigation of a possible regression.

## [0.16.3] - 2017-04-25
This release adds an stop-gap fix to the Transact method so that it never
transfers value with the CallTx is generates.

We hard-code amount = fee so that no value is transferred
regardless of fee sent. This fixes an invalid jump destination error arising
from transferring value to non-payable functions with newer versions of solidity.
By doing this we can resolve some issues with users of the v0 RPC without making
a breaking API change.

## [0.16.2] - 2017-04-20
This release finalises our accession to the Hyperledger project and updates our root package namespace to github.com/hyperledger/burrow.

It also includes a bug fix for rpc/V0 so that BroadcastTx can accept any transaction type and various pieces of internal clean-up.

## [0.16.1] - 2017-04-04
This release was an internal rename to 'Burrow' with some minor other attendant clean up.

## [0.16.0] - 2017-03-01
This is a consolidation release that fixes various bugs and improves elements
of the architecture across the Monax Platform to support a quicker release
cadence.

#### Features and improvements (among others)
- [pull-510](https://github.com/hyperledger/burrow/pull/510) upgrade consensus engine to Tendermint v0.8.0
- [pull-507](https://github.com/hyperledger/burrow/pull/507) use sha3 for snative addresses for future-proofing
- [pull-506](https://github.com/hyperledger/burrow/pull/506) alignment and consolidation for genesis and config between tooling and chains
- [pull-504](https://github.com/hyperledger/burrow/pull/504) relicense eris-db to Apache 2.0
- [pull-500](https://github.com/hyperledger/burrow/pull/500) introduce more strongly types secure native contracts
- [pull-499](https://github.com/hyperledger/burrow/pull/499) introduce word256 and remove dependency on tendermint/go-common
- [pull-493](https://github.com/hyperledger/burrow/pull/493) re-introduce GenesisTime in GenesisDoc

- Logging system overhauled based on the central logging interface of go-kit log. Configuration lacking in this release but should be in 0.16.1. Allows powerful routing, filtering, and output options for better operations and increasing the observability of an eris blockchain. More to follow.
- Genesis making is improved and moved into eris-db.
- Config templating is moved into eris-db for better synchronisation of server config between the consumer of it (eris-db) and the producers of it (eris cli and other tools).
- Some documentation updates in code and in specs.
- [pull-462](https://github.com/hyperledger/burrow/pull/499) Makefile added to capture conventions around building and testing and replicate them across different environments such as continuous integration systems.

#### Bugfixes (among others)
- [pull-516](https://github.com/hyperledger/burrow/pull/516) Organize and add unit tests for rpc/v0
- [pull-453](https://github.com/hyperledger/burrow/pull/453) Fix deserialisation for BroadcastTx on rpc/v0
- [pull-476](https://github.com/hyperledger/burrow/pull/476) patch EXTCODESIZE for native contracts as solc ^v0.4 performs a safety check for non-zero contract code
- [pull-468](https://github.com/hyperledger/burrow/pull/468) correct specifications for params on unsubscribe on rpc/tendermint
- [pull-465](https://github.com/hyperledger/burrow/pull/465) fix divergence from JSON-RPC spec for Response object
- [pull-366](https://github.com/hyperledger/burrow/pull/366) correction to circle ci script
- [pull-379](https://github.com/hyperledger/burrow/pull/379) more descriptive error message for eris-client


## [0.15.0]
This release was elided to synchronise release versions with tooling

## [0.14.0]
This release was elided to synchronise release versions with tooling

## [0.13.0]
This release was elided to synchronise release versions with tooling

## [0.12.0]
This release marks the start of Eris-DB as the full permissioned blockchain node
 of the Eris platform with the Tendermint permissioned consensus engine.
 This involved significant refactoring of almost all parts of the code,
 but provides a solid foundation to build the next generation of advanced
 permissioned smart contract blockchains.

 Many changes are under the hood but here are the main externally
 visible changes:

- Features and improvements
  - Upgrade to Tendermint 0.6.0 in-process consensus
  - Support DELEGATECALL opcode in Ethereum Virtual Machine (important for solidity library calls)
  - ARM support
  - Docker image size reduced
  - Introduction of eris-client companion library for interacting with
  eris:db
  - Improved single configuration file for all components written by eris-cm
  - Allow multiple event subscriptions from same host under rpc/tendermint


- Tool changes
  - Use glide instead of godeps for dependencies


- Testing
  - integration tests over simulated RPC calls
  - significantly improved unit tests
  - the ethereum virtual machine and the consensus engine are now top-level
  components and are exposed to continuous integration tests


- Bugfixes (incomplete list)
  - [EVM] Fix calculation of child CALL gaslimit (allowing solidity library calls to work properly)
  - [RPC/v0] Fix blocking event subscription in transactAndHold (preventing return in Javascript libraries)
  - [Blockchain] Fix getBlocks to respect block height cap.


[0.29.3]: https://github.com/hyperledger/burrow/compare/v0.29.2...v0.29.3
[0.29.2]: https://github.com/hyperledger/burrow/compare/v0.29.1...v0.29.2
[0.29.1]: https://github.com/hyperledger/burrow/compare/v0.29.0...v0.29.1
[0.29.0]: https://github.com/hyperledger/burrow/compare/v0.28.2...v0.29.0
[0.28.2]: https://github.com/hyperledger/burrow/compare/v0.28.1...v0.28.2
[0.28.1]: https://github.com/hyperledger/burrow/compare/v0.28.0...v0.28.1
[0.28.0]: https://github.com/hyperledger/burrow/compare/v0.27.0...v0.28.0
[0.27.0]: https://github.com/hyperledger/burrow/compare/v0.26.2...v0.27.0
[0.26.2]: https://github.com/hyperledger/burrow/compare/v0.26.1...v0.26.2
[0.26.1]: https://github.com/hyperledger/burrow/compare/v0.26.0...v0.26.1
[0.26.0]: https://github.com/hyperledger/burrow/compare/v0.25.1...v0.26.0
[0.25.1]: https://github.com/hyperledger/burrow/compare/v0.25.0...v0.25.1
[0.25.0]: https://github.com/hyperledger/burrow/compare/v0.24.6...v0.25.0
[0.24.6]: https://github.com/hyperledger/burrow/compare/v0.24.5...v0.24.6
[0.24.5]: https://github.com/hyperledger/burrow/compare/v0.24.4...v0.24.5
[0.24.4]: https://github.com/hyperledger/burrow/compare/v0.24.3...v0.24.4
[0.24.3]: https://github.com/hyperledger/burrow/compare/v0.24.2...v0.24.3
[0.24.2]: https://github.com/hyperledger/burrow/compare/v0.24.1...v0.24.2
[0.24.1]: https://github.com/hyperledger/burrow/compare/v0.24.0...v0.24.1
[0.24.0]: https://github.com/hyperledger/burrow/compare/v0.23.3...v0.24.0
[0.23.3]: https://github.com/hyperledger/burrow/compare/v0.23.2...v0.23.3
[0.23.2]: https://github.com/hyperledger/burrow/compare/v0.23.1...v0.23.2
[0.23.1]: https://github.com/hyperledger/burrow/compare/v0.23.0...v0.23.1
[0.23.0]: https://github.com/hyperledger/burrow/compare/v0.22.0...v0.23.0
[0.22.0]: https://github.com/hyperledger/burrow/compare/v0.21.0...v0.22.0
[0.21.0]: https://github.com/hyperledger/burrow/compare/v0.20.1...v0.21.0
[0.20.1]: https://github.com/hyperledger/burrow/compare/v0.20.0...v0.20.1
[0.20.0]: https://github.com/hyperledger/burrow/compare/v0.19.0...v0.20.0
[0.19.0]: https://github.com/hyperledger/burrow/compare/v0.18.1...v0.19.0
[0.18.1]: https://github.com/hyperledger/burrow/compare/v0.18.0...v0.18.1
[0.18.0]: https://github.com/hyperledger/burrow/compare/v0.17.1...v0.18.0
[0.17.1]: https://github.com/hyperledger/burrow/compare/v0.17.0...v0.17.1
[0.17.0]: https://github.com/hyperledger/burrow/compare/v0.16.3...v0.17.0
[0.16.3]: https://github.com/hyperledger/burrow/compare/v0.16.2...v0.16.3
[0.16.2]: https://github.com/hyperledger/burrow/compare/v0.16.1...v0.16.2
[0.16.1]: https://github.com/hyperledger/burrow/compare/v0.16.0...v0.16.1
[0.16.0]: https://github.com/hyperledger/burrow/compare/v0.15.0...v0.16.0
[0.15.0]: https://github.com/hyperledger/burrow/compare/v0.14.0...v0.15.0
[0.14.0]: https://github.com/hyperledger/burrow/compare/v0.13.0...v0.14.0
[0.13.0]: https://github.com/hyperledger/burrow/compare/v0.12.0...v0.13.0
[0.12.0]: https://github.com/hyperledger/burrow/commits/v0.12.0
