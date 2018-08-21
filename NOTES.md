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

