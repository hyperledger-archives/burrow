# Send transactions to a burrow network

#### Create a deploy file _test.yaml_
```yaml
jobs:
- name: sendTxTest1
  send:
      destination: PUT_HERE_ONE_ACCOUNT_OF_YOUR_GENESIS
      amount: 42
```

#### Send it from a node

```bash
SIGNING_ADDRESS=HERE_ONE_VALIDATOR_ADDRESS_OF_THE_GENESIS
burrow deploy --address $SIGNING_ADDRESS test.yaml
```

where you should replace the `--address` field with the `ValidatorAddress` at the top of your `burrow.toml`.

if you have updated the default burrow GRPC port, parameter `-u` `--chain` is your burrow running node ip:GRPCport.

It outputs:
```
*****Executing Job*****

Job Name                                    => defaultAddr


*****Executing Job*****

Job Name                                    => sendTxTest1


Transaction Hash                            => 41E0C13D1515F83E6FFDC5032C60682BE1F5B19A
Writing [test.output.json] to current directory
```

You can find lot of different transactions example in [jobs fixtures directory](../../tests/jobs_fixtures)

You may also start to [deploy contracts](deploy-contracts.md).
