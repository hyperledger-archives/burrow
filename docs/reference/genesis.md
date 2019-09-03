# Genesis

A blockchain network stores a series of deterministically agreed upon states indexed by a sequentially increasing number called the block height. Once a block is committed at a particular height we can always discover the total state at that height by querying more than 1/3 of the chain (in Burrow's case, and we may need to query more nodes if the ones we speak to disagree).

The state at block `n+1` is the result of applying the transactions accepted at height `n` on top of the state at height `n`. Following this induction backwards we therefore require some initial state. This initial state is called the 'genesis state'. Unlike all subsequent states that are arrived at by the application of consensus amongst validators this state is defined a priori and defines the identity of the chain that will be built on top of it. Producing the genesis document is a singular big bang event performed by a single party. It cannot occur under consensus because it defines who can participate in consensus in block 1.

## GenesisDoc
In burrow we define the genesis facts in a canonical JSON document called the `GenesisDoc`. The sha256 of this document (with whitespace, as it happens) forms the `GenesisHash` which acts as if it were the hash of a virtual zeroth block.


| Field | Purpose |
|-------|---------|
| GenesisTime | The time at which the GenesisDoc was produced - the zero time for this chain - also a source of entropy for the GenesisHash |
| ChainName | A human-readable name for the chain - also a source of entropy for the GenesisHash |
| Params | Initial parameters for the chain that control the on-chain governance process |
| GlobalPermissions | The default fall-through permissions for all accounts on the chain, see [permissions](permissions.md) |
| Accounts | The initial EVM accounts present on the chain (see below for more detail) |
| Validators | The initial validators on the chain that together will decide the value of the next state (see below for more detail) |

Here is an example `genesis.json`:
```json
{
  "GenesisTime": "2019-05-17T10:33:23.476788642Z",
  "ChainName": "BurrowChain_8B1683",
  "Params": {
    "ProposalThreshold": 3
  },
  "GlobalPermissions": {
    "Base": {
      "Perms": "send | call | createContract | createAccount | bond | name | proposal | input | batch | hasBase | hasRole",
      "SetBit": "root | send | call | createContract | createAccount | bond | name | proposal | input | batch | hasBase | setBase | unsetBase | setGlobal | hasRole | addRole | removeRole"
    }
  },
  "Accounts": [
    {
      "Address": "8E32521F19ADC32E88EACA2D23D05A3583D35A55",
      "PublicKey": {
        "CurveType": "ed25519",
        "PublicKey": "CED12BECB1BCB11F8E0268C87E4D5EE07C0224649737AAE3468373BD3F89DA1E"
      },
      "Amount": 9999999999,
      "Name": "Validator_0",
      "Permissions": {
        "Base": {
          "Perms": "bond",
          "SetBit": "bond"
        }
      }
    },
    {
      "Address": "51CA318CD3FB12697DD4FD4435C959BE025CD200",
      "PublicKey": {
        "CurveType": "ed25519",
        "PublicKey": "36F74993A6EACEB134CF02F1176EFC89E85720A5001D8D5AF46A2BCC99FBCD1E"
      },
      "Amount": 9999999999,
      "Name": "Developer_0",
      "Permissions": {
        "Base": {
          "Perms": "send | call | createContract | createAccount | name | proposal | input | hasRole | removeRole",
          "SetBit": "send | call | createContract | createAccount | name | proposal | input | hasRole | removeRole"
        }
      }
    }
  ],
  "Validators": [
    {
      "Address": "8E32521F19ADC32E88EACA2D23D05A3583D35A55",
      "PublicKey": {
        "CurveType": "ed25519",
        "PublicKey": "CED12BECB1BCB11F8E0268C87E4D5EE07C0224649737AAE3468373BD3F89DA1E"
      },
      "Amount": 9999999998,
      "Name": "Validator_0",
      "UnbondTo": [
        {
          "Address": "8E32521F19ADC32E88EACA2D23D05A3583D35A55",
          "PublicKey": {
            "CurveType": "ed25519",
            "PublicKey": "CED12BECB1BCB11F8E0268C87E4D5EE07C0224649737AAE3468373BD3F89DA1E"
          },
          "Amount": 9999999998
        }
      ]
    }
  ]
}

```
### Genesis making

TODO: burrow spec
TODO: burrow configure
