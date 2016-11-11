# mintgen
---------

Generate genesis.json files for a tendermint blockchain.

To generate a genesis.json with a single validator/account use 

```
cat /path/to/priv_validator.json | mintgen single <chain id>
```

