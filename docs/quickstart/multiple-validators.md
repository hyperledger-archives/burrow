# Multiple validators

### Configure a chain with 2 full accounts and validators
```bash
 rm -rf .burrow* .keys*
 burrow spec -f2 | burrow configure -s- --pool
 ```
### Start the network

#### Start the first node
```bash
burrow start --config=burrow000.toml
```

You will see `Blockpool has no peers` in burrow000.log.
The node has not enough validator power in order to have quorum (2/3) on the network, so it is blocked waiting for the second validator to join.

#### Start the second node

```bash
burrow start --config=burrow001.toml
```

If the connection successed, you will see empty blocks automatically created `Sending vote message` and `Finalizing commit of block with 0 txs`, you can see consensus state:
```bash
curl -s 127.0.0.1:26758/consensus
```

## Send transactions to the blockchain

You can start to [send transactions](send-transactions.md).
