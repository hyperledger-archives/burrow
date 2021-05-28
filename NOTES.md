### Changed
- [JS] Provider interface no longer depends on GRPC types to improve compatibility between versions of Burrow.js and ease of extension
- [JS] Use non-unique marker interface to indicate stream cancellation in event reducer (again for compatibility between versions and extensibility)
- [Go] Upgrade to Go 1.16

### Fixed
- [JS] Fix codegen silently swallowing collisions of abi files (renamed from .bin to .abi) and use hierarchical directory structure to further reduce chance of collision
- [JS] Just depende on @ethersproject/abi rather than entire umbrella project

### Added
- [JS] Include deployedBycode and optionally submit ABI's to Burrow's contract metadata store on deploy

