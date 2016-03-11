# Network Integration Test

Test transaction relaying with nonce-based txs

Dependencies: eris-cli, mint-client

Expects docker-machines to be already deployed as mach1, mach2, mach3, mach4.
Machine name (`mach`) can be changed at top of `setup.sh`

Start the network nodes and the local proxy: `./setup.sh`

Run the test: `./test.sh`

Use `./dev/rm.sh` to wipe the remote containers, and `eris clean` to wipe the local ones.
