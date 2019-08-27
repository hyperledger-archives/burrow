# Installation

## Binary

Download a binary from our list of [releases](https://github.com/hyperledger/burrow/releases)
and copy it to a suitable location.

## Source

[Install Go](https://golang.org/doc/install) (Version >= 1.11) and set `$GOPATH`.

```
go get github.com/hyperledger/burrow
cd $GOPATH/src/github.com/hyperledger/burrow
make build
```

This will build the `burrow` binary and put it in the `bin/` directory. It can be executed from there or put wherever is convenient.

You can also install `burrow` into `$BIN_PATH/bin` with `make install`, where `$BIN_PATH` defaults to `$HOME/go/bin`
if not set in environment.

## Docker

Each release is also tagged and pushed to [Docker Hub](https://hub.docker.com/r/hyperledger/burrow).
This can act as a direct replacement for the burrow binary, for example the following commands are equivalent:

```bash
burrow spec -v4
docker run hyperledger/burrow spec -v4
```

> Ensure to mount local volumes for secrets / configurations when running a container to prevent data loss.

## Kubernetes

Use our official [helm charts](https://github.com/helm/charts/tree/master/stable/burrow) to configure and run 
a burrow chain in your cluster.