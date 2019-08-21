# Configure & Run Burrow

## Configuration

The quick-and-dirty one-liner looks like:

```shell
# Read spec on stdin
burrow spec -p1 -f1 | burrow configure -s- > burrow.toml
```

Which translates into:

```shell
# This is a place we can store config files and burrow's working directory '.burrow'
mkdir chain_dir && cd chain_dir
burrow spec --participant-accounts=1 --full-accounts=1 > genesis-spec.json
burrow configure --genesis-spec=genesis-spec.json > burrow.toml
```

## Run Burrow
Once the `burrow.toml` has been created, we run:

```
# To select our validator address by index in the GenesisDoc
burrow start --validator=0
# Or to select based on address directly (substituting the example address below with your validator's):
burrow start --address=BE584820DC904A55449D7EB0C97607B40224B96E
```

and the logs will start streaming through.

If you would like to reset your node, you can just delete its working directory with `rm -rf .burrow`. In the context of a
multi-node chain it will resync with peers, otherwise it will restart from height 0.