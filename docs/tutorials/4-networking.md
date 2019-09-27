# Networking

So far we have only run a single validator. What happens if it stops working or loses its data?
We're much better off running multiple nodes in parallel!

## Getting Started

Let's configure a local chain with two full accounts:

```shell
burrow spec -f2 | burrow configure -s- --pool
```

You'll notice that Burrow has generated two config files instead of one, hold on to these.


## First Node

```shell
burrow start --config=burrow000.toml
```

You will see `blockpool has no peers` in the logs, this means that the node has not got enough validator power in order to have 
quorum (2/3) on the network, so it is blocked waiting for the second validator to join.

## Second Node

```shell
burrow start --config=burrow001.toml
```

If the connection succeeds, you will see empty blocks automatically created.
Look for logs such as `Sending vote message` or `Finalizing commit of block with 0 txs`.

You can also query the consensus state over our RPC with:

```shell
curl -s 127.0.0.1:26758/consensus
```