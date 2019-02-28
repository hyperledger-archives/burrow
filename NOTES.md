### Changed
- [ABI] abi.EncodeFunctionCall and AbiSpec.Pack now take a variadic ...interface{} type for function arguments rather than []string

### Fixed
- [Deploy] Binary files are now written atomically to prevent issue with dependency libraries being momentarily truncated when deploying in parallel

### Added
- [ABI] DecodeFunctionReturn re-exposed (formerly Packer then packer in 0.24.0) to make deploy API symmetrical

