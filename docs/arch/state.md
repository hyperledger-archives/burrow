# State

Burrow wraps its state in a 'mutable forest' which contains one 'commits' tree and an 'immutable forest'. The former implementation contains
commit IDs generated from all sub-trees in the 'immutable forest' thereby providing the state root hash. Each tree in our 'immutable forest'
is lazily loaded by prefix (i.e. initialized if it does not exist), returning a read/write tree. This contains immutable snapshots of its
history for reading, as well as a mutable tree for accumulating state. All trees ultimately wrap [IAVL](https://github.com/tendermint/iavl),
an (immutable) AVL+ library, persisted in [goleveldb](https://github.com/syndtr/goleveldb) - a key/value database.