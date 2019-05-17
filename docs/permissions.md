# Burrow Permissions
Burrow supports permissions via permission flags and string tags called 'roles'.
  

The flag permissions are

| Permission Flag | Capability | Purpose |
|-----------------|------------|---------|
| Root | Can issue GovernanceTxs | This an all powerful transaction type that can create accounts, change permissions, and redistribute native token and validator  power - this is often system account not controlled by a human but does not have to be | 
| Send | Can issue SendTxs | Can send native token from multiple inputs to multiple outputs provided the inputs have sufficient balance - output accounts are created dynamically so SendTx is also a means to initialise (and fund) new accounts |
| Call | Can issue CallTxs | To call EVM smart contracts |
| CreateContract | Can use the CREATE EVM opcode to create contracts programatically in the EVM | We may wish for some users to only be able to call existing contracts - sometimes the deployed contracts are our 'axiomatic' contracts, sometimes only other contracts (i.e. factories) should be able to create contracts |
| CreateAccount | Can issue actions from transactions directly or via EVM code that will create accounts in the chain's account store | We may want to prevent some users from creating other accounts as a side-effect of their actions |
| Bond | Can issue BondTxs | Allows an account with native token to convert that native token into validator power - allowing them to participate in consensus (and receive validator rewards) |
| Name | Can issue NameTxs | Allows an account to register a lease on a globally unique name with some data in a global key-value store |
| Proposal | Can issue ProposalTxs | Allows groups of accounts to vote on batches of transactions (particularly GovernanceTxs) to atomically update sets of contracts running on the network. This has particular applications in public permissioned chains where we can make proposals the only way to deploy new contracts (i.e. by not granting non-machine accounts the CreateContract permission) |
| Input | Can sign transactions | Acts as a kill-switch for specific accounts without stripping all their permissions |
| Batch | Can issue BatchTxs | Meta-transactions that a llows groups of transactions to be executed atomically within the same block |

## Initial Permissions

### Set in GenesisDoc

Either in the `GenesisDoc` field of burrow.toml or the customary separate `genesis.json` file

An example `genesis.json`:
```json
{
  "Address": "33B95296AD031ECA8F0A06B10AACF82ECB467BC1",
  "PublicKey": {
    "CurveType": "ed25519",
    "PublicKey": "2EA703B4FCC8A7186E49FF3C1D6EFBB3CAB2A97FE0A15C6A6DFC33ED87FCAB1E"
  },
  "Amount": 9999999999,
  "Name": "Developer_0",
  "Permissions": {
    "Base": {
      "Perms": "send | call | createContract | createAccount | name | proposal | input | hasRole | removeRole",
      "SetBit": "send | call | createContract | createAccount | name | proposal | input | hasRole | removeRole"
    },
    "Roles": ["custom-frog-diddly", "web-admin", "whatever"]
  }
}
```

See [genesis](genesis.md) for more details on constructing accounts and setting some useful presets
