package project

import (
	"github.com/monax/relic"
)

// Can be used to set the commit hash version of the binary at build time with:
// `go build -ldflags "-X github.com/hyperledger/burrow/project.commit=$(git rev-parse --short HEAD)" ./cmd/burrow`

var commit = ""
var date = ""

func Commit() string {
	return commit
}

func FullVersion() string {
	version := History.CurrentVersion().String()
	if commit != "" {
		return version + "+commit." + commit
	}
	return version
}

// The releases described by version string and changes, newest release first.
// The current release is taken to be the first release in the slice, and its
// version determines the single authoritative version for the next release.
//
// To cut a new release add a release to the front of this slice then run the
// release tagging script: ./scripts/tag_release.sh
var History relic.ImmutableHistory = relic.NewHistory("Hyperledger Burrow", "https://github.com/hyperledger/burrow").
	MustDeclareReleases("0.21.0 - 2018-08-21",
		`### Changed
- Upgraded to Tendermint 0.23.0
- Validator Set Power now takes Address
- RPC/TM config renamed to RPC/Info

### Added
- Burrow deploy creates devdoc
- Docker image has org.label-schema labels

### Fixed
- Upgrade to IAVL 0.10.0 and load previous versions immutably on boot - for chains with a long history > 20 minute load times could be observed because every previous root was being loaded from DB rather than lightweight version references as was intended
- Metrics server does not panic on empty block metas and recovers from other panics
`,
		"0.20.1 - 2018-08-17",
		`### Changed
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
`,
		"0.20.0 - 2018-07-24",
		`This is a major (pre-1.0.0) release that introduces the ability to change the validator set through GovTx, transaction execution history, and fuller GRPC endpoint.

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

`,
		"0.19.0 - 2018-06-26",
		`This is a major (pre-1.0.0) release that brings upgrades, safety improvements, cloud configuration, and GRPC endpoints to Burrow.

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

`,
		"0.18.1",
		`This is a minor release including:
- Introduce InputAccount param for RPC/v0 for integration in JS libs
- Resolve some issues with RPC/tm tests swallowing timeouts and not dealing with reordered events`,
		"0.18.0 - 2018-05-09",
		`This is an extremely large release in terms of lines of code changed addressing several years of technical debt. Despite this efforts were made to maintain external interfaces as much as possible and an extended period of stabilisation has taken place on develop.

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
`,

		"0.17.1",
		`Minor tweaks to docker build file`,

		"0.17.0 - 2017-09-04",
		`This is a service release with some significant ethereum/solidity compatibility improvements and new logging features. It includes:

- [Upgrade to use Tendermint v0.9.2](https://github.com/hyperledger/burrow/pull/595)
- [Implemented dynamic memory](https://github.com/hyperledger/burrow/pull/607) assumed by the EVM bytecode produce by solidity, fixing various issues.
- Logging sinks and configuration - providing a flexible mechanism for configuring log flows and outputs see [logging section in readme](https://github.com/hyperledger/burrow#logging). Various other logging enhancements.
- Fix event unsubscription
- Remove module-specific versioning
- Rename suicide to selfdestruct
- SNative tweaks

Known issues:

- SELFDESTRUCT opcode causes a panic when an account is removed. A [fix](https://github.com/hyperledger/burrow/pull/605) was produced but was [reverted](https://github.com/hyperledger/burrow/pull/636) pending investigation of a possible regression.`,

		"0.16.3 - 2017-04-25",
		`This release adds an stop-gap fix to the Transact method so that it never
transfers value with the CallTx is generates.

We hard-code amount = fee so that no value is transferred
regardless of fee sent. This fixes an invalid jump destination error arising
from transferring value to non-payable functions with newer versions of solidity.
By doing this we can resolve some issues with users of the v0 RPC without making
a breaking API change.`,

		"0.16.2 - 2017-04-20",
		`This release finalises our accession to the Hyperledger project and updates our root package namespace to github.com/hyperledger/burrow.

It also includes a bug fix for rpc/V0 so that BroadcastTx can accept any transaction type and various pieces of internal clean-up.`,

		"0.16.1 - 2017-04-04",
		`This release was an internal rename to 'Burrow' with some minor other attendant clean up.`,

		"0.16.0 - 2017-03-01",
		`This is a consolidation release that fixes various bugs and improves elements
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
`,

		"0.15.0",
		"This release was elided to synchronise release versions with tooling",

		"0.14.0",
		"This release was elided to synchronise release versions with tooling",

		"0.13.0",
		"This release was elided to synchronise release versions with tooling",

		"0.12.0",
		`This release marks the start of Eris-DB as the full permissioned blockchain node
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
`,
	)
