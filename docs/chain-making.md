# Chain Making

A Burrow network begins with the act of chain making. This process has three logical steps:

1. Generate the initial keys for the initial [participants](/docs/reference/participants.md) of the chain
2. Collect public keys and addresses from initial participants and close them into the [genesis document](reference/genesis.md) along with initial configuration for their state.
3. Boot an initial [quorum](/docs/reference/consensus.md#quorum) of network validators

The steps above can be done by hand, but we have a number sub-commands build into the `burrow` binary to help. Please see our [tutorial series](/docs/tutorials) for detailed walkthroughs.

## Generating keys

Burrow consumes its keys through our key signing interface that can be run as a standalone service with:

```shell
burrow keys server
```

This command:

- Starts a key signing daemon capable of generating new ed25519 keys, naming those keys, signing arbitrary messages, and verifying signed messages.
- Initialises a key store directory in `.keys` (by default) where private key matter is stored.

It should be noted that the GRPC service exposed by the keys server will sign _any_ inbound requests using the keys it maintains so the machine running the keys service should only allow connections from sources that are trusted to use those keys. 

## Specifying the network

## Configuring a node

## Booting a node

