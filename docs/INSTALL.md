## Installation

- [Install go](https://golang.org/doc/install) version 1.10 or above and have `$GOPATH` set

```
go get github.com/hyperledger/burrow
cd $GOPATH/src/github.com/hyperledger/burrow
make build
```

This will build the `burrow` binary and put it in the `bin/` directory. It can be executed from there or put wherever is convenient.

You can also install `burrow` into `$GOPATH/bin` with `make install_burrow`,
