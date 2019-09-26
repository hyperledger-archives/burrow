# Dump / Restore

Sometimes there are breaking changes in burrow. This provides a method for dumping an old chain, and restoring a new chain
with that state.

## Dumping Existing State

The `burrow dump` command connects to burrow node and retrieves the following:

1. The accounts (the addresses)
2. Contracts and contract storage
3. Name registry items
4. EVM Events

This can be dumped in json or go-amino format. The structure is described in (protobuf)[../protobuf/dump.proto]. By default,
it saved in go-amino, but it can be saved in json format by specify `--json`. It is also possible to dump the state at a specific
height using `--height`.

## Recreate State

You will need the `.keys` directory of the old chain, the `genesis.json` (called genesis-original in the example below)
from the old chain and the dump file (called `dump.json` here).

```shell
burrow configure -m BurrowTestRestoreNode -n "Restored Chain" -g genesis-original.json -w genesis.json --restore-dump dump.json > burrow.toml
```

Note that the chain genesis will contain an `AppHash` specific to this restore file.

## Restore Chain

This will populate the `.burrow` directory with the state.

```shell
burrow restore dump.json
```

This will create a block 0 with the restored state. Normally burrow chains start a height 1.

## Start Chain

Simply start `burrow` as you would normally.

```shell
burrow start
```

Now burrow should start making blocks at 1 as usual.