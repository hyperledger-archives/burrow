# Bonding

We can't always expect our validator set to remain the same. New participants, not established at network
formation, may wish to participate at any time.

## Getting Started

We need at least one validator to start the chain, so run the following to construct 
a genesis of two accounts with the `Bond` permission, one of which is pre-bonded:

```shell
burrow spec -v1 -r1 | burrow configure -s- --pool
```

Let's start both nodes:

```shell
burrow start --config burrow000.toml &
burrow start --config burrow001.toml &
```

Query the JSON RPC for all validators in the active set:

```shell
curl -s "localhost:26758/validators"
```

This will return the pre-bonded validator, defined in our pool.

## Joining

To have the second node bond on and produce blocks:

```shell
burrow tx --config burrow001.toml formulate bond --amount 10000 | burrow tx commit
```

Note that this will bond the current account, to bond an alternate account (which is created if it doesn't exist)
simply specific the `--source=<address>` flag in formulation:

```shell
burrow tx --config burrow001.toml formulate bond --source 8A468CC3A28A6E84ED52E433DA21D6E9ED7C1577 --amount 10000
```

It should now be in the validator set:

```shell
curl -s "localhost:26759/validators"
```

## Leaving

To unbond this validator:

```shell
burrow tx formulate unbond | burrow tx commit
```