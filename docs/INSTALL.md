## Installation

- [Install go](https://golang.org/doc/install) version 1.11 or above and have `$GOPATH` set

```
go get github.com/hyperledger/burrow
cd $GOPATH/src/github.com/hyperledger/burrow
# We need to force enable module support to build from within GOPATH (our protobuf build depends on path, otherwise any checkout location should work)
export GO111MODULE=on
make build
```

This will build the `burrow` binary and put it in the `bin/` directory. It can be executed from there or put wherever is convenient.

You can also install `burrow` into `$GOPATH/bin` with `make install_burrow`,
