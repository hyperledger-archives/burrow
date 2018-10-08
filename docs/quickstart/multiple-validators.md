# Multiple validators

### Configure a chain with 2 full accounts and validators
```bash
 rm -rf .burrow* .keys*
 burrow spec -f2 | burrow configure -s- > .burrow_init.toml
```

### Configure 2 validator nodes config files
From the generated `.burrow_init.toml `file, create new files for each node, and change the content, example:

#### Validator 1 node `.burrow_val0.toml` modified line from `.burrow_init.toml`

```toml
[Tendermint]
  Seeds = ""
  SeedMode = false
  PersistentPeers = ""
  ListenAddress = "tcp://0.0.0.0:20000"
  Moniker = "val_node_0"
  TendermintRoot = ".burrow_node0"

[Execution]

[Keys]
  GRPCServiceEnabled = false
  AllowBadFilePermissions = true
  RemoteAddress = ""
  KeysDirectory = ".keys"

[RPC]
  [RPC.Info]
    Enabled = true
    ListenAddress = "tcp://127.0.0.1:20001"
  [RPC.Profiler]
    Enabled = false
  [RPC.GRPC]
    Enabled = true
    ListenAddress = "127.0.0.1:20002"
  [RPC.Metrics]
    Enabled = false
```

#### Validator 2 node `.burrow_val1.toml` modified line from `.burrow_init.toml`

```toml
[Tendermint]
  Seeds = ""
  SeedMode = false
  PersistentPeers = "PUT_HERE_NODE_0_ID@LISTEN_EXTERNAL_ADDRESS"
  ListenAddress = "tcp://0.0.0.0:30000"
  Moniker = "val_node_1"
  TendermintRoot = ".burrow_node1"

[Execution]

[Keys]
  GRPCServiceEnabled = false
  AllowBadFilePermissions = true
  RemoteAddress = ""
  KeysDirectory = ".keys"

[RPC]
  [RPC.Info]
    Enabled = true
    ListenAddress = "tcp://127.0.0.1:30001"
  [RPC.Profiler]
    Enabled = false
  [RPC.GRPC]
    Enabled = true
    ListenAddress = "127.0.0.1:30002"
  [RPC.Metrics]
    Enabled = false
```

Node 0 will be defined as persistent peer of node 1.
Persistent peers are people you want to be constantly connected with.

### Start the network

#### Start the first node
```bash
burrow start --validator-index=0 --config=.burrow_val0.toml
```

You will see `Blockpool has no peers` in console logs.
The node has not enough validator power in order to have quorum (2/3) on the network, so it is blocked waiting for the second validator to join.

#### Find the first node identifier and address

Configure second node to persistently connect to the first node.

```bash
NODE_0_URL=`curl -s 127.0.0.1:20001/network | jq -r '.result.ThisNode | [.ID, .ListenAddress] | join("@") | ascii_downcase'`
sed -i s%PUT_HERE_NODE_0_ID@LISTEN_EXTERNAL_ADDRESS%${NODE_0_URL}% .burrow_val1.toml
```

#### Start the second node

```bash
burrow start --validator-index=1 --config=.burrow_val1.toml
```

If the connection successed, you will see empty blocks automatically created `Sending vote message` and `Finalizing commit of block with 0 txs`, you can see consensus state:
```bash
curl -s 127.0.0.1:20001/consensus
```

#### Disable Tendermint strict address if required
If you face `Cannot add non-routable address` message in logs, it means your listen address is not routable for tendermint.

You can disable this check by modified default tendermint/config.go:46 and rebuild burrow:
```go
conf.P2P.AddrBookStrict = false
```

Or explicitly set an external routable address, for example on node 0, if you have a net interface on `172.217.19.227`:
```toml
[Tendermint]
  ExternalAddress = "172.217.19.227:20000"
```

Update `PersistentPeers` property of node 1 with this new address.

## Send transactions to the blockchain

You can start to [send transactions](send-transactions.md).