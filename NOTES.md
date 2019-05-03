### Changed
- [Config] Split ListenAddress into ListenHost and ListenPort to ease parsing in the Helm charts
- [CLI] Burrow restore now always fails if state is detected but can be made --silent
- [CLI] No dump client timeout by default
- [Deploy] Reduced the default logging level to trace instead of info
- [Build] Switched to Go modules

### Fixed
- [Keys] Resolved an issue where the keyStore wasn't built when using the remote keys client.
- [Deploy] Fix nil dereference in query error path, check constructor args in BuildJob

