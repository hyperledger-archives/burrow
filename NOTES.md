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


