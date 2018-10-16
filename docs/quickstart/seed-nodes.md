# Network with seed nodes

## What is a seed node

According to [Tendermint documentation](https://tendermint.com/docs/tendermint-core/using-tendermint.html#seed):
>A seed node is a node who relays the addresses of other peers which they know
of. These nodes constantly crawl the network to try to get more peers. The
addresses which the seed node relays get saved into a local address book. Once
these are in the address book, you will connect to those addresses directly.
Basically the seed nodes job is just to relay everyones addresses. You won't
connect to seed nodes once you have received enough addresses, so typically you
only need them on the first start. The seed node will immediately disconnect
from you after sending you some addresses.

### Seed mode
SeedMode can be enabled on a node.

If a node is in seed mode it will accept inbound connections, share its address book, then hang up.
Seeds modes will do a bit of gossip but not that usefully.
Any type of node can be referenced as a 'Seeds' in the config, whether or not another node considers this node as a seed is independent of whether this node is in 'seed mode'.
These are different concepts:
> You are free to use a non-seed-mode node as a seed.

You do not want to have too many seeds in your network (because they just keep hanging up on other nodes once they've spread their wild oats), but they are useful for accelerating peer exchange (of addresses).

### Persistent peers
Persistent peers are peers that you want to connect of regardless of the heuristics and churn dynamics built into the p2p switch.
Ordinarily you would not stay connected to a particular peer forever, and you would not indefinitely redial, but you will for a persistent peer.

## Configure network

In this quick start, we will few create validator nodes which does not know each other from network point of view.
A seed node will crawl the network and relay everyones addresses.

### Configure chain

```bash
rm -rf .burrow* .keys*
burrow spec --full-accounts=3 | burrow configure -s- > .burrow_init.toml
```

### Generate one additional key in another local store for seed node
```bash
burrow spec -f1 | burrow configure --keysdir=.keys_seed -s- > /dev/null
```

### Make 3 validator nodes and one seed node config files
From the generated `.burrow_init.toml `file, create new files for each node, and change the content, example:

#### Seed node `.burrow_seed.toml` modified line from `.burrow_init.toml`
```toml
[Tendermint]
  SeedMode = true
  ListenAddress = "tcp://0.0.0.0:10000"
  Moniker = "seed_node_0"
  TendermintRoot = ".burrow_seed_0"

[Execution]

[Keys]
  GRPCServiceEnabled = false
  AllowBadFilePermissions = true
  RemoteAddress = ""
  KeysDirectory = ".keys_seed"

[RPC]
  [RPC.Info]
    Enabled = true
    ListenAddress = "tcp://127.0.0.1:10001"
  [RPC.Profiler]
    Enabled = false
  [RPC.GRPC]
    Enabled = false
  [RPC.Metrics]
    Enabled = false
```

#### Validator 1 node `.burrow_val0.toml` modified line from `.burrow_init.toml`

```toml
[Tendermint]
  Seeds = "PUT_HERE_SEED_NODE_ID@LISTEN_EXTERNAL_ADDRESS"
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
  Seeds = "PUT_HERE_SEED_NODE_ID@LISTEN_EXTERNAL_ADDRESS"
  SeedMode = false
  PersistentPeers = ""
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


#### Validator 3 node `.burrow_val2.toml` modified line from `.burrow_init.toml`

```toml
[Tendermint]
  Seeds = "PUT_HERE_SEED_NODE_ID@LISTEN_EXTERNAL_ADDRESS"
  SeedMode = false
  PersistentPeers = ""
  ListenAddress = "tcp://0.0.0.0:40000"
  Moniker = "val_node_2"
  TendermintRoot = ".burrow_node2"

[Execution]

[Keys]
  GRPCServiceEnabled = false
  AllowBadFilePermissions = true
  RemoteAddress = ""
  KeysDirectory = ".keys"

[RPC]
  [RPC.Info]
    Enabled = true
    ListenAddress = "tcp://127.0.0.1:40001"
  [RPC.Profiler]
    Enabled = false
  [RPC.GRPC]
    Enabled = true
    ListenAddress = "127.0.0.1:40002"
  [RPC.Metrics]
    Enabled = false
```

#### Start the seed node
```bash
burrow start --validator-address=`basename .keys_seed/data/* .json` --config=.burrow_seed.toml  > .burrow_seed.log 2>&1 &
```

#### Find seed node external address
Tendermint requires strict and routable address (not loopback, local etc), you can find the listen address with this command:
```bash
SEED_URL=`curl -s 127.0.0.1:10001/network | jq -r '.result.ThisNode | [.ID, .ListenAddress] | join("@") | ascii_downcase'`
echo $SEED_URL
```

#### Configure other node to connect to seed node
Update other nodes with that seed address:
```bash
sed -i s%PUT_HERE_SEED_NODE_ID@LISTEN_EXTERNAL_ADDRESS%${SEED_URL}% .burrow_val0.toml
sed -i s%PUT_HERE_SEED_NODE_ID@LISTEN_EXTERNAL_ADDRESS%${SEED_URL}% .burrow_val1.toml
sed -i s%PUT_HERE_SEED_NODE_ID@LISTEN_EXTERNAL_ADDRESS%${SEED_URL}% .burrow_val2.toml
```

#### Start validator nodes
```bash
burrow start --validator-index=0 --config=.burrow_val0.toml  > .burrow_val0.log 2>&1 &
burrow start --validator-index=1 --config=.burrow_val1.toml  > .burrow_val1.log 2>&1 &
burrow start --validator-index=2 --config=.burrow_val2.toml  > .burrow_val2.log 2>&1 &
```

Nodes will connect to seed node and request addresses, then they will connect to each other and start submitting and voting on blocks.

Check network status, validators nodes are connected to each others:
```bash
curl -s 127.0.0.1:40001/network | jq -r '.result.peers[].node_info.moniker'
val_node_0
val_node_1
```

You can monitor consensus and current blockchain height from the node info websockets:
```bash
curl -s 127.0.0.1:20001/consensus | jq -r '.result.round_state.height'
```

At the moment, there is an [issue](https://github.com/tendermint/tendermint/issues/2092) opened in Tendermint with seedMode.

Disable seed mode on the seed node and see how it affects the peers network:

```toml
[Tendermint]
  SeedMode = false
```

Clear nodes folder (Note it will restart the chain from the genesis block):
```bash
killall burrow
rm -rf .burrow_node0 .burrow_node1 .burrow_node2 .burrow_seed_0
```

Restart all nodes, then check network status (Validator 3 is now connected to all peers, included seed node):
```bash
curl -s 127.0.0.1:40001/network | jq -r '.result.peers[].node_info.moniker'
seed_node_0
val_node_0
val_node_1
```
