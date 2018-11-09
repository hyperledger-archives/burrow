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

