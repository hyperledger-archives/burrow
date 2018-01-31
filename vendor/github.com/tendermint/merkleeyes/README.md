# merkleeyes

[![CircleCI](https://circleci.com/gh/tendermint/merkleeyes.svg?style=svg)](https://circleci.com/gh/tendermint/merkleeyes)

A simple [ABCI application](http://github.com/tendermint/abci) serving a [merkle-tree key-value store](http://github.com/tendermint/merkleeyes/iavl) 

# Use

Merkleeyes allows inserts and removes by key, and queries by key or index.
Inserts and removes happen through the `DeliverTx` message, while queries happen through the `Query` message.
`CheckTx` simply mirrors `DeliverTx`.

# Formatting

## Byte arrays

Byte-array `B` is serialized to `Encode(B)` as follows:

```
Len(B) := Big-Endian encoded length of B
Encode(B) = Len(Len(B)) | Len(B) | B
```

So if `B = "eric"`, then `Encode(B) = 0x010465726963`

## Transactions

There are two types of transaction, each associated with a type-byte and a list of arguments:

```
Set			0x01		Key, Value
Remove			0x02		Key
```

A transaction consists of the type-byte concatenated with the encoded arguments.

For instance, to insert a key-value pair, you would submit `01 | Encode(key) | Encode(value)`. 
Thus, a transaction inserting the key-value pair `(eric, clapton)` would look like:

```
0x010104657269630107636c6170746f6e
```


Here's a session from the [abci-cli](https://tendermint.com/intro/getting-started/first-abci):

```
> deliver_tx 0x010104657269630107636c6170746f6e

> commit
-> data: ��N��٢ek�X�!a��
-> data.hex: 978A4ED807D617D9A2651C6B0EC9588D2161C9E0

> query 0x65726963                  
-> height: 2
-> key: eric
-> key.hex: 65726963
-> value: clapton
-> value.hex: 636C6170746F6E
```

# Poem

```
writing down, my checksum
waiting for the, data to come
no need to pray for integrity
thats cuz I use, a merkle tree

grab the root, with a quick hash run
if the hash works out,
it must have been done

theres no need, for trust to arise
thanks to the crypto
now that I can merkleyes

take that data, merklize
ye, I merklize ...

then the truth, begins to shine
the inverse of a hash, you will never find
and as I watch, the dataset grow
producing a proof, is never slow

Where do I find, the will to hash
How do I teach it?
It doesn't pay in cash
Bitcoin, here, I've realized
Thats what I need now,
cuz real currencies merklize
-EB
```
