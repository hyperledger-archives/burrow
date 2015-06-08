# Eris DB

Eris DB allows remote access to its functionality over http and websockets. It currently supports JSON-RPC (over both http and websockets), and REST-like http. There is also javascript bindings available in the [erisdb-js](TODO) library.

<a name="json-rpc"></a>
## JSON RPC 2.0

The default endpoints for JSON-RPC (2.0) is `/rpc` for http based, and `/socketrpc` for websockets. The namespace for the JSON-RPC service is `erisdb`. 


### Objects

##### Errors

```
PARSE_ERROR      = -32700
INVALID_REQUEST  = -32600
METHOD_NOT_FOUND = -32601
INVALID_PARAMS   = -32602
INTERNAL_ERROR   = -32603
```

#####Request
```
{
	jsonrpc: <string>
	method:  <string>
	params:  <Object>
	id:      <string>
}
```

#####Response
```
{
	jsonrpc: <string>
	id:      <string>
	result:  <Object>
	error:   <Error>
}
```

#####Error
```
{
    code:    <number>
    message: <string>
}
```
Id can be any string value. Parameters are named, and wrapped in objects. Also, params, result and error params may be `null`.

#####Example

Request: 
```
{
	jsonrpc: "2.0", 
	method: "erisdb.getAccount", 
	params: {address: "37236DF251AB70022B1DA351F08A20FB52443E37"}, 
	id="25"
}
```

Response:
```
{
    address: "37236DF251AB70022B1DA351F08A20FB52443E37",
    pub_key: null,
    sequence: 0,
    balance: 110000000000,
    code: "",
    storage_root: ""
}
```

<a name="rest-like"></a>
## REST-like api

The REST-like web-api provides the typical endpoint structure i.e. endpoints are resources, GET params in path, and queries used only for filtering. It is not fully compatible with normal REST and probably won't be. This is partly because some GET requests can contain sizable input so POST is used instead. There are also some modeling issues but those will most likely be resolved before 1.0.


<a name="formatting-conventions"></a>
##Common objects and formatting

This section contains some common objects and explanations of how they behave.
  
###Numbers and strings

Numbers are always numbers, and never strings. This is different from Ethereum where currency values are so high they need string representations. The only thing hex strings are used for is to represent byte arrays. 

Hex strings are never prefixed. 

#####Examples

```
"some_number_field" : 5892,
"another_number_field" : 0x52
"hex_string" : "37236DF251AB70022B1DA351F08A20FB52443E37"
```

###Keys and addresses

Public and Private keys in JSON data are either null, or on the form: `[type, hex]`, where `type` is the [public](https://github.com/tendermint/tendermint/blob/master/account/pub_key.go), or [private](https://github.com/tendermint/tendermint/blob/master/account/pub_key.go) key type, and `hex` is the hex-string representation of the key bytes.

* A `public address` is a 20 byte hex string.

* A `public key` is a 32 byte hex string.

* A `private key` is a 64 byte hex string.

#####WARNING

**When using a client-server setup, do NOT send public keys over non-secure connections. The only time this is fine is during development when the keys are nothing but test data and does not protect anything of value. Normally they should either be kept locally and used to sign transactions locally, held on the server where the blockchain client is running, or be passed over secure channels.**

#####Examples

A public address: `"37236DF251AB70022B1DA351F08A20FB52443E37"`

The corresponding Ed25519 public key: `[1, "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"]`

The corresponding Ed25519 private key: `[1, "6B72D45EB65F619F11CE580C8CAED9E0BADC774E9C9C334687A65DCBAD2C4151CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"]`


<a name="the-transaction-types"></a>
###The transaction types

These are the types of transactions:

####SendTx
```
{
	inputs:  [<TxInput>]
	outputs: [<TxOutput>]
}
```

####CallTx
```
{
	input:     <TxInput>
	address:   <string>
	gas_limit: <number>
	fee:       <number>
	data:      <string>
}
```

####NameTx
```
{
	input: <TxInput>
	name:  <string>
	data:  <string>
	fee:   <number>
}
```

####BondTx
```
{
	pub_key:   <PubKey>
	signature: <string>
	inputs:    [<TxInput>]
	unbond_to: [<TxOutput>]
}
```

####UnbondTx
```
{
	address:   <string>
	height:    <number>
	signature: <string>
}
```

####RebondTx
```
{
	address:   <string>
	height:    <number>
	signature: <string>
}
```

####DupeoutTx
```
{
	address: <string>
	vote_a:  <Vote>
	vote_b:  <Vote>
}
```

These are the support types that are referenced in the transactions:

####TxInput
```
{
	address:   <string>
	amount:    <number>
	sequence:  <number>
	signature: <string>
	pub_key:   <string>
}
```

####TxOutput
```
{
	address: <string>
	amount:  <number>
}
```

####Vote

```
{
	height:     <number>
	type:       <number>
	block_hash: <string>
	block_parts: {
		total: <number>
		hash:  <string>
	}
	signature: <string>
}
```

<a name="event-system"></a>
##Event system


####Contract code

TODO

<a name="methods"></a>
##Methods

###Accounts 
| Name | RPC method name | REST method | REST endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [GetAccounts](#get-accounts) | erisdb.getAccounts | GET | `/accounts` |
| [GetAccount](#get-account) | erisdb.getAccount | GET | `/accounts/:address` |
| [GetStorage](#get-storage) | erisdb.getStorage | GET | `/accounts/:address/storage` |
| [GetStorageAt](#get-storage-at) | erisdb.getStorageAt | GET | `/accounts/:address/storage/:key` |

###Blockchain
| Name | RPC method name | REST method | REST endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [GetBlockchainInfo](#get-blockchain-info) | erisdb.getBlockchainInfo | GET | `/blockchain` |
| [GetChainId](#get-chain-id) | erisdb.getChainId | GET | `/blockchain/chain_id` |
| [GetGenesisHash](#get-genesis-hash) | erisdb.getGenesisHash | GET | `/blockchain/genesis_hash` |
| [GetLatestBlockHeight](#get-latest-block-height) | erisdb.getLatestBlockHeight | GET | `/blockchain/latest_block/height` |
| [GetLatestBlock](#get-latest-block) | erisdb.getLatestBlock | GET | `/blockchain/latest_block` |
| [GetBlocks](#get-blocks) | erisdb.getBlocks | GET | `/blockchain/blocks` |
| [GetBlock](#get-block) | erisdb.getBlock | GET | `/blockchain/blocks/:height` |

###Consensus
| Name | RPC method name | REST method | REST endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [GetConsensusState](#get-consensus-state) | erisdb.getConsensusState | GET | `/consensus` |
| [GetValidators](#get-validators) | erisdb.getValidators | GET | `/consensus/validators` |

###Events
| Name | RPC method name | REST method | REST endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [EventSubscribe](#event-subscribe) | erisdb.eventSubscribe | POST | `/event_subs` |
| [EventUnsubscribe](#event-unsubscribe) | erisdb.eventUnsubscribe | DELETE | `/event_subs/:id` |
| [EventPoll](#event-poll) | erisdb.eventPoll | GET | `/event_subs/:id` |

###Network
| Name | RPC method name | REST method | REST endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [GetNetworkInfo](#get-network-info) | erisdb.getNetworkInfo | GET | `/network` |
| [GetMoniker](#get-moniker) | erisdb.getMoniker | GET | `/network/moniker` |
| [GetChainId](#get-chain-id) | erisdb.getChainId | GET | `/network/chain_id` |
| [IsListening](#is-listening) | erisdb.isListening | GET | `/network/listening` |
| [GetListeners](#get-listeners) | erisdb.getListeners | GET | `/network/listeners` |
| [GetPeers](#get-peers) | erisdb.getPeers | GET | `/network/peers` |
| [GetPeer](#get-peer) | erisdb.getPeer | GET | `/network/peer/:address` |

###Transactions
| Name | RPC method name | REST method | REST endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [BroadcastTx](#broadcast-tx) | erisdb.broadcastTx | POST | `/txpool` |
| [GetUnconfirmedTxs](#get-unconfirmed-txs) | erisdb.broadcastTx | GET | `/txpool` |

###Code execution
| Name | RPC method name | REST method | REST endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [Call](#call) | erisdb.call | POST | `/calls` |
| [CallCode](#call-code) | erisdb.callCode | POST | `/calls/code` |


####Unsafe
| Name | RPC method name | REST method | REST endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [SignTx](#sign-tx) | erisdb.signTx | POST | `/unsafe/tx_signer` |
| [Transact](#transact) | erisdb.transact | POST | `/unsafe/txpool` |
| [GenPrivAccount](#gen-priv-account) | erisdb.genPrivAccount | GET | `/unsafe/pa_generator` |

Here are the catagories.

* [Accounts](#accounts)
* [BlockChain](#blockchain)
* [Consensus](#consensus)
* [Events](#events)
* [Network](#network)
* [Transactions](#transactions)
* [Code Execution (calls)](#calls)
* [Unsafe](#unsafe)

In the case of **JSON-RPC**, the parameters are wrapped in a request object, and the return value is wrapped in a response object.

In the case of **REST**, the params (and query) is provided in the url of the request. If it's a POST, PATCH or PUT request, the parameter object should be written to the body of the request in JSON form. It is normally the same object as would be the params in the corresponding JSON-RPC request.

**Unsafe** is methods that require a private key to be sent either to or from the client, and should therefore be used only during development/testing, or with extreme care. They may be phased out entirely.

<a name="accounts"></a>
###Accounts 

***

<a name="get-accounts"></a>
####GetAccounts 

Get accounts will return a list of accounts. If no filtering is used, it will return all existing accounts.

#####HTTP

Method: GET 

Endpoint: `/accounts`

Search terms: 

`

#####JSON-RPC

Method: `erisdb.getAccounts`

Parameter:

```
{
	filters: [<FilterData>]
}
```


#####Return value
``` 
{
	accounts: [<Account>]
}
```

#####Additional info

See GetAccount below for more info on the `Account` object.

See the section on [Filters](#filters) for info on the `FilterData` object.

***

<a name="get-account"></a>
####GetAccount

Get an account by its address.

#####HTTP

Method: GET 

Endpoint: `/accounts/:address`

Params: The public `address` as a hex string.


#####JSON-RPC

Method: `erisdb.getAccount`

Parameter:

```
{
	address: <string>
}
```

#####Return value
``` 
{
	address:      <string>
	pub_key:      <PubKey>
	sequence:     <number>
	balance:      <number>
	code:         <string>
	storage_root: <string>
}
```

`address` is a public address.
`pub_key` is a public key.

#####Additional info

Sequence is sometimes referred to as the "nonce".

There are two types of objects used to represent accounts, one is public accounts (like the one here), the other is private accounts, which only holds information about an accounts address, public and private key.

***

<a name="get-storage"></a>
####GetStorage 

Get the complete storage of a contract account. Non-contract accounts has no storage. 

NOTE: This is mainly used for debugging. In most cases the storage of an account would be accessed via public accessor functions defined in the contracts ABI.

#####HTTP

Method: GET 

Endpoint: `/accounts/:address/storage`

Params: The public `address` as a hex string.


#####JSON-RPC

Method: `erisdb.getStorage`

Parameter:

```
{
	address: <string>
}
```

#####Return value
``` 
{
	storage_root:  <string>
	storage_items: [<StorageItem>]
}
```

`storage_root` is a public address.
See `GetStorageAt` below for more info on the `StorageItem` object.

***

<a name="get-storage-at"></a>
####GetStorageAt

Get a particular entry in the storage of a contract account. Non-contract accounts has no storage. 

NOTE: This is mainly used for debugging. In most cases the storage of an account would be accessed via public accessor functions defined in the contracts ABI.

#####HTTP

Method: GET

Endpoint: `/accounts/:address/storage/:key`

Params: The public `address` as a hex string, and the `key` as a hex string.

#####JSON-RPC

Method: `erisdb.getStorageAt`

Parameter:

```
{
	address: <string>
	key:     <string>
}
```

#####Return value
``` 
{
	key:   <string>
	value: <string>
}
```

Both `key` and `value` are hex strings.

***

<a name="blockchain"></a>
###Blockchain

***

<a name="get-blockchain-info"></a>
####GetBlockchainInfo

Get the current state of the blockchain. This includes things like chain-id and latest block height. There are individual getters for all fields as well. 

#####HTTP

Method: GET

Endpoint: `/blockchain`

#####JSON-RPC

Method: `erisdb.getBlockchainInfo`

Parameter: -

#####Return value
``` 
{
	chain_id:            <string>
	genesis_hash:        <string>
	latest_block:        <BlockMeta>
	latest_block_height: <number> 
}
```

#####Additional info

`chain_id` is the name of the chain.
`genesis_hash` is a 32 byte hex-string. It is the hash of the genesis block, which is the first block on the chain.
`latest_block` contains block metadata for the latest block. See the [GetBlock](#get-block) method for more info.
`latest_block_height` is the height of the latest block, and thus also the height of the entire chain.

The block *height* is sometimes referred to as the block *number*.

See [GetBlock](#get-block) for more info on the `BlockMeta` type.

***

<a name="get-chain-id"></a>
####GetChainId

Get the chain id.

#####HTTP

Method: GET

Endpoint: `/blockchain/chain_id`

#####JSON-RPC

Method: `erisdb.getChainId`

Parameter: -

#####Return value
``` 
{
	chain_id:            <string>
}
```

***

<a name="get-genesis-hash"></a>
####GetGenesisHash

Get the genesis hash. This is a 32 byte hex-string representation of the hash of the genesis block. The genesis block is the first block on the chain.

#####HTTP

Method: GET

Endpoint: `/blockchain/genesis_hash`

#####JSON-RPC

Method: `erisdb.getGenesisHash`

Parameter: -

#####Return value
``` 
{
	genesis_hash:        <string> 
}
```

***

<a name="get-latest-block-height"></a>
####GetLatestBlockHeight

Get the height of the latest block. This would also be the height of the entire chain.

#####HTTP

Method: GET

Endpoint: `/blockchain/latest_block/height`

#####JSON-RPC

Method: `erisdb.getLatestBlockHeight`

Parameter: -

#####Return value
``` 
{
	latest_block_height: <number> 
}
```

***

<a name="get-latest-block"></a>
####GetLatestBlock

Gets the block that was added to the chain most recently. 

#####HTTP

Method: GET

Endpoint: `/blockchain/latest_block`

#####JSON-RPC

Method: `erisdb.getLatestBlock`

Parameter: -

#####Return value
``` 
{
	latest_block:        <BlockMeta> 
}
```

#####Additional info

See [GetBlock](#get-block) for more info on the `BlockMeta` type. 

***

<a name="get-blocks"></a>
####GetBlocks

Get a series of blocks from the chain. 

#####HTTP

Method: GET

Endpoint: `/blockchain/blocks`

#####JSON-RPC

Method: `erisdb.getBlocks`

Parameter: 

```
{
	filters: [<FilterData>]
}
```

#####Return value

``` 
{
	min_height:  <number>
	max_height:  <number>
	block_metas: [<BlockMeta>]
}
```

The `BlockMeta` object:

```
{
	hash: <string>
	header: {
		chain_id:        <string>
		height:          <number>
		time:            <string>
		fees:            <number>
		num_txs:         <number>
		last_block_hash: <string>
		last_block_parts: {
			total: <number>
			hash:  <string>
		}
		state_hash: <string>
	}
	parts: {
		total: <number>
		hash:  <string>
	}
}
```

#####Additional info

TODO

See the section on [Filters](#filters) for info on the `FilterData` object.

`min_height` and `max_height` is the two actual values used for min and max height when fetching the blocks. The reason they are included is because the heights might have been modified, like for example when the blockchain height is lower then the max height provided in the query.

See [GetBlock](#get-block) for more info on the `BlockMeta` type. 

***

<a name="get-block"></a>
####GetBlock

Get the block at the given height. 

#####HTTP

Method: GET

Endpoint: `/blockchain/block/:number`

#####JSON-RPC

Method: `erisdb.getBlock`

Parameter:

``` 
{
	height: <number> 
}
```

#####Return value
``` 
{
	
	header: {
		chain_id:        <string>
		height:          <number>
		time:            <string>
		fees:            <number>
		num_txs:         <number>
		last_block_hash: <string>
		last_block_parts: {
			total: <number>
			hash:  <string>
		}
		state_hash: <string>
	}
	validation: {
		commits: [<Commit>]
		TODO those other two.
	}
	data: {
		txs: [<Tx>]
		TODO that other field.
	}
}
```

The `Commit` object:

```
{
	address:   <string>
	round:     <number>
	signature: <string>
}
```

#####Additional info

TODO

See [The transaction types](#the-transaction-types) for more info on the `Tx` types. 

***

<a name="consensus"></a>
###Consensus

***

<a name="get-consensus-state"></a>
####GetConsensusState

Get the current consensus state. 

#####HTTP

Method: GET

Endpoint: `/consensus`

#####JSON-RPC

Method: `erisdb.getConsensusState`

Parameter: -

#####Return value

``` 
{
	height:      <number>
	round:       <number>
	step:        <number>
	start_time:  <string>
	commit_time: <string>
	validators:  [<Validator>]
	proposal: {
		height: <number>
		round:  <number>
		block_parts: {
			total: <number>
			hash:  <string>
		}
		pol_parts: {
			total: <number>
			hash:  <string>
		}
		signature: <string>
	}
}
```

#####Additional info

TODO

See the GetValidators method right below for info about the `Validator` object.

***

<a name="get-validators"></a>
####GetValidators

Get the validators. 

#####HTTP

Method: GET

Endpoint: `/consensus/validators`

#####JSON-RPC

Method: `erisdb.getValidators`

Parameter: -

#####Return value

``` 
{
	block_height:         <number>
	bonded_validators:    [<Validator>]
	unbonding_validators: [<Validator>]
}
```

The `Validator` object:

```
{
	address:            <string>
	pub_key:            <PubKey>
	bon_height:         <number>
	unbond_height:      <number>
	last_commit_height: <number>
	voting_power:       <number>
	accum:              <number>
}
```

#####Additional info

TODO

***

<a name="events"></a>
###Events

***

<a name="event-subscribe"></a>
####EventSubscribe

Subscribe to a given type of event.  

#####HTTP

Method: POST

Endpoint: `/event_subs/`

Body: See JSON-RPC parameter.

#####JSON-RPC

Method: `erisdb.eventSubscribe`

Parameter: 

``` 
{
	event_id: <string>
}
```

#####Return value

``` 
{
	sub_id: <string>
}
```

#####Additional info

For more information about events and the event system, see the [Event system](#event-system) section.

***

<a name="event-unsubscribe"></a>
####EventUnubscribe

Unsubscribe to an event type.  

#####HTTP

Method: DELETE

Endpoint: `/event_subs/:id`

#####JSON-RPC

Method: `erisdb.eventUnsubscribe`

Parameter: -

#####Return value

``` 
{
	result: <bool>
}
```

#####Additional info

For more information about events and the event system, see the [Event system](#event-system) section.

***

<a name="event-poll"></a>
####EventPoll

Poll a subscription. Note this cannot be done if using websockets, because then the events will be passed automatically over the socket.

#####HTTP

Method: GET

Endpoint: `/event_subs/:id`

#####JSON-RPC

Method: `erisdb.eventPoll`

#####Return value

``` 
{
	events: [<Event>]
}
```

#####Additional info

For more information about events and the event system, see the [Event system](#event-system) section. This includes info about the `Event` object.

***

<a name="network"></a>
###Network

***

<a name="get-network-info"></a>
####GetNetworkInfo

Get the network information. This includes the blockchain client moniker, peer data, and other things.

#####HTTP

Method: GET

Endpoint: `/network`

#####JSON-RPC

Method: `erisdb.getNetworkInfo`

Parameters: -

#####Return value

``` 
{
	moniker: <string>
	listening: <boolean>
	listeners: [<string>]
	peers: [<Peer>]
}
```

#####Additional info

`moniker` is a moniker for the node.
`listening` is a check if the node is listening for connections.
`listeners` is a list of active listeners.
`peers` is a list of peers.

See [GetPeer](#get-peer) for info on the `Peer` object.

***

<a name="get-moniker"></a>
####GetMoniker

Get the node moniker, or nickname.

#####HTTP

Method: GET

Endpoint: `/network/moniker`

#####JSON-RPC

Method: `erisdb.getMoniker`

Parameters: -

#####Return value

``` 
{
	moniker: <string>
}
```

***

<a name="is-listening"></a>
####IsListening

Check whether or not the node is listening for connections.

#####HTTP

Method: GET

Endpoint: `/network/listening`

#####JSON-RPC

Method: `erisdb.isListening`

Parameters: -

#####Return value

``` 
{
	listening: <boolean>
}
```

***

<a name="get-listeners"></a>
####GetListeners

Get a list of all active listeners.

#####HTTP

Method: GET

Endpoint: `/network/listeners`

#####JSON-RPC

Method: `erisdb.getListeners`

Parameters: -

#####Return value

``` 
{
	listeners: [<string>]
}
```

***

<a name="get-peers"></a>
####GetPeers

Get a list of all peers.

#####HTTP

Method: GET

Endpoint: `/network/peers`

#####JSON-RPC

Method: `erisdb.getPeers`

Parameters: -

#####Return value

``` 
{
	peers: [<Peer>]
}
```

See [GetPeer](#get-peer) below for info on the `Peer` object.

***

<a name="get-peer"></a>
####GetPeer

Get the peer with the given IP address.

#####HTTP

Method: GET

Endpoint: `/network/peer/:address`

#####JSON-RPC

Method: `erisdb.getPeer`

Parameters: 

``` 
{
	address: <string>
}
```

#####Return value

This is the peer object.

``` 
{
	is_outbound: <boolean>
	moniker:     <string>
	chain_id:    <string>
	version:     <string>
	host:        <string>
	p2p_port:    <number>
	rpc_port:    <number>
}
```


#####Additional info

TODO

***

<a name="transactions"></a>
###Transactions 

***

<a name="BroadcastTx"></a>
####BroadcastTx

Broadcast a given (signed) transaction to the node. It will be added to the tx pool if there are no issues, and if it is accepted by all validators it will eventually be committed to a block. 

#####HTTP

Method: POST

Endpoint: `/txpool`

Body:

``` 
<Tx>
```

#####JSON-RPC

Method: `erisdb.BroadcastTx`

Parameters: 

``` 
<Tx>
```

#####Return value

``` 
{
	tx_hash:          <string>
	creates_contract: <number>
	contract_addr:    <string>
}
```

#####Additional info

`tx_hash` is the hash of the transaction (think digest), and can be used to reference it.

`creates_contract` is set to `1` if a contract was created, otherwise it is 0.

If a contract was created, then `contract_addr` will contain the address. NOTE: This is no guarantee that the contract will actually be commited to the chain. This response is returned upon broadcasting, not when the transaction has been committed to a block.

See [The transaction types](#the-transaction-types) for more info on the `Tx` types. 

***

<a name="get-unconfirmed-txs"></a>
####GetUnconfirmedTxs

Get a list of transactions currently residing in the transaction pool. These have been admitted to the pool, but has not yet been committed.

#####HTTP

Method: GET

Endpoint: `/txpool`

#####JSON-RPC

Method: `erisdb.getUnconfirmedTxs`

Parameters: -

#####Return value

``` 
{
	txs: [<Tx>]
}
```


#####Additional info

See [The transaction types](#the-transaction-types) for more info on the `Tx` types. 

***

<a name="calls"></a>
###Code execution (calls) 

***

<a name="Call"></a>
####Call

Call a given (contract) account to execute its code with the given in-data. 

#####HTTP

Method: POST

Endpoint: `/calls`

Body: See JSON-RPC parameter.

#####JSON-RPC

Method: `erisdb.call`

Parameters: 

``` 
{
	address: <string>
	data: <string>
}
```

#####Return value

``` 
{
	return:   <string>
	gas_used: <number>
}
```

#####Additional info

`data` is a string of data formatted in accordance with the [contract ABI](TODO).

***

<a name="CallCode"></a>
####Call

Pass contract code and tx data to the node and have it executed in the virtual machine. This is mostly a dev feature.

#####HTTP

Method: POST

Endpoint: `/calls/code`

Body: See JSON-RPC parameter.

#####JSON-RPC

Method: `erisdb.callCode`

Parameters: 

``` 
{
	code: <string>
	data: <string>
}
```

#####Return value

``` 
{
	return: <string>
	gas_used: <number>
}
```

#####Additional info

`code` is a hex-string representation of compiled contract code.
`data` is a string of data formatted in accordance with the [contract ABI](TODO).

***

<a name="unsafe"></a>
###Unsafe 

These methods are unsafe because they require that a private key is either transmitted or received. They are supposed to be used mostly in development/debugging, and should normally not be used in a production environment.

***

<a name="SignTx"></a>
####SignTx

Send an unsigned transaction to the node for signing. 

#####HTTP

Method: POST

Endpoint: `/unsafe/tx_signer`

Body:

``` 
<Tx>
```

#####JSON-RPC

Method: `erisdb.SignTx`

Parameters: 

``` 
<Tx>
```

#####Return value

The same transaction but signed.

#####Additional info

See [The transaction types](#the-transaction-types) for more info on the `Tx` types. 

***

<a name="Transact"></a>
####Transact

Convenience method for sending a transaction "old Ethereum dev style". It will do the following things:

* Use the private key to create a private account object (i.e. generate public key and address).
* Use the other parameters to create a `CallTx` object.
* Sign the transaction.
* Broadcast the transaction.

#####HTTP

Method: POST

Endpoint: `/unsafe/txpool`

Body: See JSON-RPC parameters.

#####JSON-RPC

Method: `erisdb.SignTx`

Parameters: 

``` 
{
	priv_key:  <PrivKey>
	data:      <string>
	address:   <string>
	fee:       <number>
	gas_limit: <number>
}
```

#####Return value

The same as with BroadcastTx:

``` 
{
	tx_hash:          <string>
	creates_contract: <number>
	contract_addr:    <string>
}
```

#####Additional info

See [The transaction types](#the-transaction-types) for more info on the `CallTx` type. 

***

<a name="GenPrivAccount"></a>
####GenPrivAccount

Convenience method for generating a `PrivAccount` object, which contains a private key and the corresponding public key and address.

#####HTTP

Method: POST

Endpoint: `/unsafe/pa_generator`
#####JSON-RPC

Method: `erisdb.genPrivAccount`

Parameters: -

#####Return value

``` 
{
	address: <string>
	pub_key: <PubKey>
	priv_key: <PrivKey>
}
```

#####Additional info

TODO fix endpoint and method.

Again - This is unsafe. Be warned.

***

<a name="filters"></a>
##Filters

Filters are used in searches. The filter query structure is similar to that of the [Github api (v3)](https://developer.github.com/v3/search/).

###JSON-RPC

Filters are added as objects in the request parameter. Methods that supports filtering includes an array of filters somewhere in their params object.

Filter:
```
{
    field: <string>
    op:    <string>
    value: <*> 
}
```

* The `field` must be one that is supported by the collection items in question.
* The `op` is a relational operation `[>, <, >=, <=, ==, !=]`. Different fields supports different subsets.
* The `value` is the value to match against. It is always a string.
* Range queries are done simply by adding two filters - one for the minimum value and one for the maximum.

#####Examples

We want an account filter that only includes accounts that has code in them (i.e. contract accounts):

```
{
    field: "code"
    op: "!="
    value: "" 
}
```

We want an account filter that only includes accounts with a balance less then 1000:

```
{
    field: "balance"
    op: "<"
    value: "1000"
}
```

We want an account filter that only includes accounts with a balance higher then 0, but less then 1000.

```
{
    field: "balance"
    op: ">"
    value: "0"
}
```

```
{
    field: "balance"
    op: "<"
    value: "1000"
}
```

The field `code` is supported by accounts. It allows for the `==` and `!=` operators. The value `""` means the empty hex string.

If we wanted only non-contract accounts then we would have used the same object but changed it to `op: "=="`. 

###HTTP Queries

The structure of a normal query is: `q=field:statement+field2:statement2+ ... `.

*`q` means it's a query.
*`field` is the field name.
*`:` is a separator.
* `statement` is normally on the form `[op]value` where `op` is a relational operator, and `value` a value, e.g. `balance:>=5` or `language:==golang`. There is also support for [range queries](https://help.github.com/articles/search-syntax/): `*..*`, where `*` is a number. If on the left of `..` it means minimum value, and if it's on the right it means maximum value. `height:*..55`. There is only the non-quoted version as of now.

NOTE: URL encoding applies as usual. Omitting it here for clarity.

If you leave `op` out it will default to equals (`==`).

`value` may be left empty if the underlying type for that field is a string. This means if `code` is a supported string type,  `/accounts?q=code:%3D%3D` would check if the code field is empty. We could also use inferred `==` here, meaning this query would be equivalent: `/accounts?q=code:`. 

#####Example

We want to use the same filter as in the JSON version; one that finds all contract accounts.

`http://localhost:1337/accounts?q=code:%21%3D` (code != "")

One that finds those with balance less then 1000:

`http://localhost:1337/accounts?q=balance:%3C1000` (balance < 1000)

One that finds those with balance more then 0, but less then 1000.

`http://localhost:1337/accounts?q=balance:0..1000`

One that finds contract accounts with a balance less then 1000:

`http://localhost:1337/accounts?q=code:%21%3D+balance:0..1000`

###Supported types

Right now you can only filter searches when getting blocks and accounts. The system will expand to handle a lot more before version 1.0.

#### Accounts

#####Code

Field name: `code`

Ops: `==`, `!=`

Field type: hex-string

Underlying type: byte-array
 
`http://localhost:1337/accounts?q=code:%3D%3Dabababab` (code == "abababab")

#####Balance

Field name: `balance`

Ops: All

Field type: integer-string.

Underlying type: uint64

`http://localhost:1337/accounts?q=balance:0..1000`

#### Blocks

#####Height

Field name: `height`

Ops: `<`, `>`, `<=`, `>=`, `==` (`!=` is not supported because disjoint result sets are not allowed) 

Field type: integer string

Underlying type: uint

`http://localhost:1337/blockchain/blocks?q=height:0..10`