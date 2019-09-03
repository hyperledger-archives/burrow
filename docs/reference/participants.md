# Accounts

Accounts are identified by an address - this is the EVM's 20-byte native identifier and is encoded in hex in the GenesisDoc. The public key is redundant but included for reference. An account's address is defined as the first 20 bytes of the sha256 hash of the account's public key so an address can be derived from a public key but not the other way around.

The `Amount` is an initial amount of Burrow's native token. This native token can serve multiple purposes:

- Funding an account by sending token implicitly creates that account (if it does not already exist). Only accounts that have been created by having value transferred to them (via a `SendTx` or an EVM `Call`) can themselves act as input accounts (e.g. in a `CallTx`).
- The native token can be used as a value-holding token on a Burrow sidechain in the same was that eth acts as a value token on public Ethereum.
- Native token can be be converted via a [BondTx](/docs/reference/transactions.md#bondtx) into validator voting power.

Accounts can also hold EVM or WASM code that is initialised on account creation (and thereafter cannot be changed - except for by [GovTx](/docs/reference/transactions.md#govtx)). You can create a contract by sending a [CallTx](/docs/reference/transactions.md#calltx) using

## Validators

Validators are the nodes running on the network that are permitted to participant block proposal and voting according to the Tendermint consensus algorithm. Each validator is identified by it public key (ed25519) and can also be described by a corresponding 20-byte address. A validator is assigned a `Power` this determines the relative power of each of its votes and how often it will be rotated into position of block proposer.

