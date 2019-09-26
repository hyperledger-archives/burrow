# Basics

You can spin up a single node chain with:

```shell
burrow spec -v1 | burrow configure -s- | burrow start -c-
```

## Configuration

The quick-and-dirty one-liner looks like:

```shell
# Read spec on stdin
burrow spec -p1 -f1 | burrow configure -s- > burrow.toml
```

Which translates into:

```shell
burrow spec --participant-accounts=1 --full-accounts=1 > genesis-spec.json
burrow configure --genesis-spec=genesis-spec.json > burrow.toml
```

> You might want to run this in a clean directory to avoid overwriting any previous spec or config.

## Running

Once the `burrow.toml` has been created, we run:

```
# To select our validator address by index in the GenesisDoc
burrow start --validator=0
# Or to select based on address directly (substituting the example address below with your validator's):
burrow start --address=BE584820DC904A55449D7EB0C97607B40224B96E
```

If you would like to reset your node, you can just delete its working directory with `rm -rf .burrow`. 
In the context of a multi-node chain it will resync with peers, otherwise it will restart from height 0.

## Keys

Burrow consumes its keys through our key signing interface that can be run as a standalone service with:

```shell
burrow keys server
```

This command starts a key signing daemon capable of generating new ed25519 and secp256k1 keys, naming those keys, signing arbitrary messages, and verifying signed messages.
It also initializes a key store directory in `.keys` (by default) where private key matter is stored.

It should be noted that the GRPC service exposed by the keys server will sign _any_ inbound requests using the keys it maintains so the machine running the keys service should only allow connections from sources that are trusted to use those keys. 