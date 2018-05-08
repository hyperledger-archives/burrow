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

