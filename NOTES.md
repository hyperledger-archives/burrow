### Fixed
- Upgrade to IAVL 0.10.0 and load previous versions immutably on boot - for chains with a long history > 20 minute load times could be observed because every previous root was being loaded from DB rather than lightweight version references as was intended
