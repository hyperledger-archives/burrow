# Dump / Restore

Sometimes there are breaking changes in burrow. This provides a method for dumping an old chain, and restoring a new chain
with that state.

## Dumping existing state

The `burrow dump` command connects to burrow node and retrieves the following:

1. The accounts (the addresses)
2. Contracts and contract storage
3. Name registry items
4. EVM Events

This can be dumped in json or go-amino format. The structure is described in (protobuf)[../protobuf/dump.proto]. By default,
it saved in go-amino, but it can be saved in json format by specify `--json`. It is also possible to dump the state at a specific
height using `--height`.

## Creating a new chain genesis with state

So you will need the `.keys` directory of the old chain, the `genesis.json` (called genesis-original in the example below)
from the old chain and the dump file (called `dump.json` here).

```bash
burrow configure -m BurrowTestRestoreNode -n "Restored Chain" -g genesis-original.json -w genesis.json --restore-dump dump.json > burrow.toml
```

Note that the chain genesis will contain an `AppHash` specific to this restore file.

## Restart the chain with the state

This will populate the `.burrow` directory with the state.

```bash
burrow restore dump.json
```

This will create a block 0 with the restored state. Normally burrow chains start a height 1.

## Run the new chain

Simply start `burrow` as you would normally.

```bash
burrow start
```

Now burrow should start making blocks at 1 as usual.