### Fixed
- [Dump] Stop TxStack EventStream consumer from rejecting events from dump/restored chain because they lack tx Envelopes (as they are intended to to keep dump format minimal)
- [Genesis] Fix hash instability introduced by accidentally removing omitempty from AppHash in genesis

### Added
- [Vent] Implement throttling on Ethereum Vent consumer via --max-request-rate=<requests / time base> flag to 'vent start'

