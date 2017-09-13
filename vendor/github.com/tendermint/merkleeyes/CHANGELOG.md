# Changelog

## 0.2.4 (June 27, 2017)

FEATURES:
- support `--trace` and `--log_level` flags like other binaries

IMPROVEMENTS:
- support standard viper functionality like other binaries
- `MerkleEyesApp` uses a logger not just `fmt.Println`
- logger configured properly in process

## 0.2.3 (June 21, 2017)

FEATURES:
- [app] `Info()` now implements the ABCI handshake, allowing the app to recover from the latest state after crash
- [app] `State.Hash()` returns the latest root of the DeliverTx tree
- [iavl] `IAVLTree.BatchSet(key, value)` adds to the current batch in the NodeDB for atomic writes

IMPROVEMENTS:
- Better README
- More comments in code
- Better testing from shell scripts

## 0.2.2 (June 5, 2017)

BUG FIXES:
- Actually start the Merkleeyes server

## 0.2.1 (June 2, 2017)

IMPROVEMENTS:
- Add version number to the source code
- Update dependencies

## 0.2.0 (May 18, 2017)

Merge in the IAVL tree from `go-merkle` and update import paths for new `tmlibs`

## 0.1.0 (March 6, 2017)

Initial release
