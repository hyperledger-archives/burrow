## Installation

- [Install go](https://golang.org/doc/install) version 1.10 or above and have `$GOPATH` set

```
go get github.com/hyperledger/burrow
cd $GOPATH/src/github.com/hyperledger/burrow
make build
```

This will build the `burrow` and `burrow-client` binaries and put them in the `bin/` directory. They can be executed from there or put wherever is convenient.

You can also install `burrow` into `$GOPATH/bin` with `make install_burrow`,