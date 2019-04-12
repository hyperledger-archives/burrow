# Add Validator

For this example, make sure you have the latest `burrow` binary and JSON parsing tool `jq` installed.

First, let's start a network with two running validators:

```bash
burrow spec -f2 | burrow configure -s- --pool --separate-genesis-doc=genesis.json
burrow start --config=burrow000.toml &
burrow start --config=burrow001.toml &
```

Next, fetch a persistent peer and the validator address of the first node: 

```bash
PERSISTENT_PEER=$(cat burrow000.toml | grep PersistentPeers | cut -d \" -f2)
OLD_VALIDATOR=$(cat burrow000.toml | grep ValidatorAddress | cut -d \" -f2)
```

Let's generate the config and keys for a new validator account. As this node will be joining an existing network we won't need the GenesisDoc, but we will need to give it the persistent peer address we obtained above. Unless you want to do these steps manually, please use the following commands:

```bash
burrow spec -v1 | burrow configure -s- --json > burrow-new.json
NEW_VALIDATOR=$(jq -r '.GenesisDoc.Accounts[0].PublicKey.PublicKey' burrow-new.json)
jq 'del(.GenesisDoc)' burrow-new.json | jq ".Tendermint.PersistentPeers=\"$PERSISTENT_PEER\"" | jq '.RPC.Info.Enabled=false' | jq '.RPC.GRPC.Enabled=false' | jq '.Tendermint.ListenPort="25565"' > burrow003.json 
```

Copy the following script into `deploy.yaml`:

```yaml
jobs:
- name: InitialTotalPower
  query-vals:
    field: "Set.TotalPower"

- name: AddValidator
  update-account:
    target: NEW_VALIDATOR
    power: 232322

- name: CheckAdded
  query-vals:
    field: "Set.${AddValidator.address}.Power"

- name: AssertPowerNonZero
  assert:
    key: $CheckAdded
    relation: gt
    val: 0

- name: AssertPowerEqual
  assert:
    key: $CheckAdded
    relation: eq
    val: $AddValidator.power
```

If you haven't already, swap in the public key for your new validator and run the deployment to give the account stake:

```bash
sed -i "s/NEW_VALIDATOR/$NEW_VALIDATOR/" deploy.yaml
burrow deploy -u 127.0.0.1:10997 --mempool-signing=true --address=$OLD_VALIDATOR deploy.yaml
```

If this returns successfully, you'll be able to see that the new validator is now in the running set:

```bash
curl -s 127.0.0.1:26758/consensus
```

Let's start the new validator and watch it catch up:

```bash
burrow start --config=burrow003.json --genesis=genesis.json
```