# Transactions

Burrow supports a number of [transactions](reference/transactions.md) which denote a unit of computation.
The easiest way to experiment is with our `burrow tx` command, but please checkout the [deployment guide](deploy.md)
for more advanced usage.

## Getting Started

Let's start a chain with one validator to process blocks and two participant accounts:

```shell
burrow spec -v1 -p2 | burrow configure -s- > burrow.toml
burrow start -v0 &
```

Make a note of the two participant addresses generated in the `burrow.toml`.

## Send Token

Let's formulate a transaction to send funds from one account to another.
Given our two addresses created above, set `$SENDER` and `$RECIPIENT` respectively.
We'll also need to designate an amount of native token available from our sender.

```shell
burrow tx formulate send -s $SENDER -t $RECIPIENT -a $AMOUNT > tx.json
```

To send this transaction to your local node and subsequently the chain (if running more than one validator),
pipe the output above through the following command:

```shell
burrow tx commit --file tx.json
```