# State

State store by burrow is written our EVM or WASM engines on executing bytecode provided by [`CallTx`](/docs/reference/transactions.md#CallTx) transactions that have been accepted into the blockchain. The state is stored through the following generic interface:

```go
type StorageSetter interface {
	// Store a 32-byte value at key for the account at address, setting to Zero256 removes the key
	SetStorage(address crypto.Address, key binary.Word256, value []byte) error
}
```

The raw data stored according depends on a schema determined by the execution engine and contract in question, in the case of the EVM this is described by the ABI generated when a contract is compiled.

Burrow stores its state in an authenticated key-value data structure - a merkle tree. It has the following features:

- We store a separate complete version of all core state at each height - this gives us the ability to rewind instantly to any height.
- We are able to provide inclusion proofs for any element of state (not currently exposed by our RPC interfaces).
- State has a single unified state root hash that almost surely guarantees identity of state by comparison between state root hashes

## Data structure

Burrow stores its core state the `Forext` which is implemented as a merkle tree of commit objects for individual sub-trees thereby providing the state root hash. Each tree in our forest is lazily loaded by prefix (i.e. initialized if it does not exist), returning a read/write tree. This contains immutable snapshots of its history for reading, as well as a mutable tree for accumulating state. All trees ultimately wrap [IAVL](https://github.com/tendermint/iavl), an (immutable) AVL+ library, persisted in [goleveldb](https://github.com/syndtr/goleveldb) - a key/value database.

### Index and derivable data

Alongside our core data we have additional data that can be derived from (such as indices) or is peripheral to (such as contract metadata). Since we can generally detect if these are incorrect or regenerate them we store them in a plain non-authenticated key-value storage called the `Plain`

### Relationship with Tendermint state

Tendermint also uses merkle trees to store raw block and transaction data. Tendermint blocks close in our state root hash as the `AppHash` thereby creating a merkle graph that conveys the authenticated data structure property to our application state. 
