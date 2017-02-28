> Hell is other people's packages

While we wait for working package management in go we need a way to make
maintaining the glide.lock by hand less painful.

To interactively add a package run from the root:

```bash
go run ./util/hell/cmd/hell/main.go get --interactive github.com/tendermint/tendermint
```

