# Bonding Validators

As Burrow runs on Tendermint, it supports the notion of bonding validators.

## Example

We need at least one validator to start the chain, so run the following to construct 
a genesis of two accounts with the `Bond` permission, one of which is pre-bonded:

```bash
burrow spec -v1 -r1 | burrow configure -s- --pool
```

Let's start the both nodes:

```bash
burrow start --config burrow000.toml &
burrow start --config burrow001.toml &
```

Query the JSON RPC for all validators in the active set:

```bash
curl -s "localhost:26758/validators"
```

This will return the pre-bonded validator, defined in our pool.

To have the second node bond on and produce blocks:

```bash
burrow tx --config burrow001.toml formulate bond --amount 10000 | burrow tx commit
```

Note that this will bond the current account, to bond an alternate account (which is created if it doesn't exist)
simply specific the `--source=<address>` flag in formulation:

```bash
burrow tx --config burrow001.toml formulate bond --source 8A468CC3A28A6E84ED52E433DA21D6E9ED7C1577 --amount 10000
```

It should now be in the validator set:

```bash
curl -s "localhost:26759/validators"
```

To unbond this validator:

```bash
burrow tx formulate unbond | burrow tx commit
```