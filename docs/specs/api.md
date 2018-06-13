# Burrow web APIs (draft)

### for burrow version 0.11.x

Burrow allows remote access to its functionality over http and websocket. It currently supports [JSON-RPC 2.0](http://www.jsonrpc.org/specification), and REST-like http. There is also javascript bindings available in the [burrow-js](https://github.com/hyperledger/burrow.js) library.

## TOC

- [HTTP Requests](#http-requests)
- [JSON-RPC 2.0](#json-rpc)
- [REST-like HTTP](#rest-like)
- [Common objects and formatting](#formatting-conventions)
- [Event-system](#event-system)
- [Methods](#methods)
- [NameReg](#namereg)
- [Filters](#queries-filters)

<a name="http-requests"></a>
## HTTP Requests

The only data format supported is JSON. All post requests need to use `Content-Type: application/json`. The charset flag is not supported (json is utf-8 encoded by default).

<a name="json-rpc"></a>
## JSON RPC 2.0

The default endpoints for JSON-RPC (2.0) are `/rpc` for http based and `/socketrpc` for websocket. The namespace for the JSON-RPC service is `burrow`.

It does not yet support notifications or batched requests.

### Objects

##### Errors

```
PARSE_ERROR      = -32700
INVALID_REQUEST  = -32600
METHOD_NOT_FOUND = -32601
INVALID_PARAMS   = -32602
INTERNAL_ERROR   = -32603
```

##### Request

```
{
	jsonrpc: <string>
	method:  <string>
	params:  <Object>
	id:      <string>
}
```

##### Response

```
{
	jsonrpc: <string>
	id:      <string>
	result:  <Object>
	error:   <Error>
}
```

##### Error

```
{
    code:    <number>
    message: <string>
}
```

Id can be any string value. Parameters are named, and wrapped in objects. Also, the params, result and error elements may be `null`.

##### Example

Request:

```
{
	jsonrpc: "2.0",
	method: "burrow.getAccount",
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
## REST-like HTTP

The REST-like API provides the typical endpoint structure--i.e. endpoints are named as resources, parameters can be put in the path, and queries are used for filtering and such. It is not fully compatible with REST partly because some GET requests can contain sizable input, so POST is used instead. There are also some modeling issues, but those will most likely be resolved before version 1.0.

<a name="formatting-conventions"></a>
## Common objects and formatting

This section contains some common objects and explanations of how they work.

### Numbers and strings

Numbers are always numbers and never strings. This is different from Ethereum where currency values are so high they need string representations. The only thing hex strings are used for is to represent byte arrays. Hex strings are never prefixed.

##### Examples

```
"some_number_field" : 5892,
"another_number_field" : 0x52
"hex_string" : "37236DF251AB70022B1DA351F08A20FB52443E37"
```

### Keys and addresses

Public and Private keys in JSON data are either null, or on the form: `[type, hex]`, where `type` is the [public](https://github.com/tendermint/tendermint/blob/master/account/pub_key.go) or [private](https://github.com/tendermint/tendermint/blob/master/account/pub_key.go) key type, and `hex` is the hex-string representation of the key bytes.

- A `public address` is a 20 byte hex string.
- A `public key` is a 32 byte hex string.
- A `private key` is a 64 byte hex string.

##### WARNING

**When using a client-server setup, do NOT send private keys over non-secure connections. The only time this is fine is during development when the keys are nothing but test data and does not protect anything of value. Normally they should either be kept locally and used to sign transactions locally, held on the server where the blockchain client is running, or be passed over secure channels.**

##### Examples

A public address: `"37236DF251AB70022B1DA351F08A20FB52443E37"`

The corresponding Ed25519 public key: `[1, "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"]`

The corresponding Ed25519 private key: `[1, "6B72D45EB65F619F11CE580C8CAED9E0BADC774E9C9C334687A65DCBAD2C4151CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"]`

<a name="the-transaction-types"></a>
### The transaction types

These are the types of transactions. Note that in DApp programming you would only use the `CallTx`, and maybe `NameTx`.

#### SendTx

```
{
	inputs:  [<TxInput>]
	outputs: [<TxOutput>]
}
```

#### CallTx

```
{
	input:     <TxInput>
	address:   <string>
	gas_limit: <number>
	fee:       <number>
	data:      <string>
}
```

#### NameTx

```
{
	input:  <TxInput>
	name:   <string>
	data:   <string>
	amount: <number>
	fee:    <number>
}
```

#### BondTx

```
{
	pub_key:   <PubKey>
	signature: <string>
	inputs:    [<TxInput>]
	unbond_to: [<TxOutput>]
}
```

#### UnbondTx

```
{
	address:   <string>
	height:    <number>
	signature: <string>
}
```

#### RebondTx

```
{
	address:   <string>
	height:    <number>
	signature: <string>
}
```

These are the support types that are referenced in the transactions:

#### TxInput

```
{
	address:   <string>
	amount:    <number>
	sequence:  <number>
	signature: <string>
	pub_key:   <string>
}
```

#### TxOutput

```
{
	address: <string>
	amount:  <number>
}
```

#### Vote

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
## Event system

Tendermint events can be subscribed to regardless of what connection type is used. There are three methods for this:

- [EventSubscribe](#event-subscribe) is used to subscribe to a given event, using an event-id string as argument. The response will contain a `subscription ID`, which can be used to close down the subscription later, or poll for new events if using HTTP. More on event-ids below.
- [EventUnsubscribe](#event-unsubscribe) is used to unsubscribe to an event. It requires you to pass the `subscription ID` as an argument.
- [EventPoll](#event-poll) is used to get all the events that have accumulated since the last time the subscription was polled. It takes the `subscription ID` as a parameter. NOTE: This only works over HTTP. Websocket connections will automatically receive events as they happen. They are sent as regular JSON-RPC 2.0 responses with the `subscriber ID` as response id.

There is another slight difference between polling and websocket, and that is the data you receive. If using sockets, it will always be one event at a time, whereas polling will give you an array of events.

### Event types

These are the type of events you can subscribe to.

The "Account" events are triggered when someone transacts with the given account, and can be used to keep track of account activity.

NewBlock and Fork happen when a new block is committed or when a fork occurs, respectively.

The other events are directly related to consensus. You can find out more about the Tendermint consensus system in the Tendermint [white paper](http://tendermint.com/docs/tendermint.pdf). There is also information in the consensus [sources](https://github.com/tendermint/tendermint/blob/master/consensus/state.go), although a normal user would not be concerned with the consensus mechanisms, but would mostly just listen to account-events and perhaps block-events.

#### Account Input

This notifies you when an account is receiving input.

Event ID: `Acc/<address>/Input`

Example: `Acc/B4F9DA82738D37A1D83AD2CDD0C0D3CBA76EA4E7/Input` will subscribe to input events from the account with address: B4F9DA82738D37A1D83AD2CDD0C0D3CBA76EA4E7.

Event object:

```
{
	tx:        <Tx>
	return:    <string>
	exception: <string>
}
```

#### Account Output

This notifies you when an account is yielding output.

Event ID: `Acc/<address>/Output`

Example: `Acc/B4F9DA82738D37A1D83AD2CDD0C0D3CBA76EA4E7/Output` will subscribe to output events from the account with address: B4F9DA82738D37A1D83AD2CDD0C0D3CBA76EA4E7.

Event object:

```
<Tx>
```

#### Account Call

This notifies you when an account is the target of a call. This event is emitted when `CallTx`s (transactions) that target the given account has been finalized. It is possible to listen to this event when creating new contracts as well; it will fire when the transaction is committed (or not, in which case the 'exception' field will explain why it failed).

**NOTE: The naming here is a bit unfortunate. Ethereum uses 'transaction' for (state-changing) transactions to a contract account, and 'call' for read-only calls like those used for accessor functions and such. Tendermint, on the other hand, uses 'CallTx' for a transaction made to a contract account, since it calls the code in that contract, and refers to these simply as 'calls'. Read-only calls is normally referred to as 'simulated calls'.**

Event ID: `Acc/<address>/Call`

Example: `Acc/B4F9DA82738D37A1D83AD2CDD0C0D3CBA76EA4E7/Call` will subscribe to events from the account with address: B4F9DA82738D37A1D83AD2CDD0C0D3CBA76EA4E7.


```
{
	call_data: {
		caller: <string>
    	callee: <string>
    	data:   <string>
    	value:  <number>
    	gas:    <number>
	}
	origin:     <string>
	tx_id:      <string>
	return:     <string>
	exception:  <string>
}
```

#### Log

This notifies you when the VM fires a log-event. This happens for example when a solidity event is fired.

Event ID: `Log/<address>`

Example: `Log/B4F9DA82738D37A1D83AD2CDD0C0D3CBA76EA4E7/Input` will subscribe to all log events from the account with address: B4F9DA82738D37A1D83AD2CDD0C0D3CBA76EA4E7.

type Log struct {
	Address Word256
	Topics  []Word256
	Data    []byte
	Height  uint64
}

Event object:

```
{
	address: <string>
	topics:  []<string>
	data:    <string>
	height   <number>
}
```

`address` is the address of the account that created the log event.

`topics` is the parameters listed as topics. In a (named) Solidity event they would be the hash of the event name, followed by each param with the `indexed` modifier.

`data` is data. In a Solidity event these would be the params without the `indexed` modifier.

`height` is the current block-height.

#### New Block

This notifies you when a new block is committed.

Event ID: `NewBlock`

Event object:

```
<Block>
```

#### Fork

This notifies you when a fork event happens.

Event ID: `Fork`

Event object:

TODO
```
<Block>
```

#### Bond

This notifies you when a bond event happens.

Event ID: `Bond`

Event object:

```
<Tx>
```

#### Unbond

This notifies you when an unbond event happens.

Event ID: `Unbond`

Event object:

```
<Tx>
```

#### Rebond

This notifies you when a rebond event happens.

Event ID: `Rebond`

Event object:

```
<Tx>
```

<a name="namereg">
### Name-registry

The name-registry is a built-in key-value store that allows you to store bulk data in a different storage. It is currently regulated by the use of Tendermint tokens. The cost of storing some `Data` in the name-registry is this:

```
TotalCost = Cost*NumberOfBlocks

Cost = CostPerBlock*CostPerByte*(length(Data) + 32)

CostPerBlock = 1

CostPerByte = 1

length(Data) = the number of bytes in 'Data'.
```

To pay this cost you use the `amount` field in the namereg transaction. If you want to store a 3 kb document for 10 blocks, the total cost would be `1*1*(3000 + 32)*10 = 30320` tendermint tokens.

See the [TransactNameReg](#transact-name-reg) method for more info about adding entries to the name-registry and the methods in the [Name-registry](#name-registry) for accessing them.

<a name="methods"></a>
## Methods

### Accounts
| Name | RPC method name | HTTP method | HTTP endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [GetAccounts](#get-accounts) | burrow.getAccounts | GET | `/accounts` |
| [GetAccount](#get-account) | burrow.getAccount | GET | `/accounts/:address` |
| [GetStorage](#get-storage) | burrow.getStorage | GET | `/accounts/:address/storage` |
| [GetStorageAt](#get-storage-at) | burrow.getStorageAt | GET | `/accounts/:address/storage/:key` |

### Blockchain
| Name | RPC method name | HTTP method | HTTP endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [GetBlockchainInfo](#get-blockchain-info) | burrow.getBlockchainInfo | GET | `/blockchain` |
| [GetChainId](#get-chain-id) | burrow.getChainId | GET | `/blockchain/chain_id` |
| [GetGenesisHash](#get-genesis-hash) | burrow.getGenesisHash | GET | `/blockchain/genesis_hash` |
| [GetLatestBlockHeight](#get-latest-block-height) | burrow.getLatestBlockHeight | GET | `/blockchain/latest_block/height` |
| [GetLatestBlock](#get-latest-block) | burrow.getLatestBlock | GET | `/blockchain/latest_block` |
| [GetBlocks](#get-blocks) | burrow.getBlocks | GET | `/blockchain/blocks` |
| [GetBlock](#get-block) | burrow.getBlock | GET | `/blockchain/blocks/:height` |

### Consensus
| Name | RPC method name | HTTP method | HTTP endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [GetConsensusState](#get-consensus-state) | burrow.getConsensusState | GET | `/consensus` |
| [GetValidators](#get-validators) | burrow.getValidators | GET | `/consensus/validators` |

### Events
| Name | RPC method name | HTTP method | HTTP endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [EventSubscribe](#event-subscribe) | burrow.eventSubscribe | POST | `/event_subs` |
| [EventUnsubscribe](#event-unsubscribe) | burrow.eventUnsubscribe | DELETE | `/event_subs/:id` |
| [EventPoll](#event-poll) | burrow.eventPoll | GET | `/event_subs/:id` |

### Name-registry
| Name | RPC method name | HTTP method | HTTP endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [GetNameRegEntry](#get-namereg-entry) | burrow.getNameRegEntry | GET | `/namereg/:key` |
| [GetNameRegEntries](#get-namereg-entries) | burrow.getNameRegEntries | GET | `/namereg` |

### Network
| Name | RPC method name | HTTP method | HTTP endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [GetNetworkInfo](#get-network-info) | burrow.getNetworkInfo | GET | `/network` |
| [GetClientVersion](#get-client-version) | burrow.getClientVersion | GET | `/network/client_version` |
| [GetMoniker](#get-moniker) | burrow.getMoniker | GET | `/network/moniker` |
| [IsListening](#is-listening) | burrow.isListening | GET | `/network/listening` |
| [GetListeners](#get-listeners) | burrow.getListeners | GET | `/network/listeners` |
| [GetPeers](#get-peers) | burrow.getPeers | GET | `/network/peers` |
| [GetPeer](#get-peer) | burrow.getPeer | GET | `/network/peer/:address` |

NOTE: Get peer is not fully implemented.

### Transactions
| Name | RPC method name | HTTP method | HTTP endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [BroadcastTx](#broadcast-tx) | burrow.broadcastTx | POST | `/txpool` |
| [GetUnconfirmedTxs](#get-unconfirmed-txs) | burrow.getUnconfirmedTxs | GET | `/txpool` |

### Code execution
| Name | RPC method name | HTTP method | HTTP endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [Call](#call) | burrow.call | POST | `/calls` |
| [CallCode](#call-code) | burrow.callCode | POST | `/codecalls` |

#### Unsafe
| Name | RPC method name | HTTP method | HTTP endpoint |
| :--- | :-------------- | :---------: | :------------ |
| [Transact](#transact) | burrow.transact | POST | `/unsafe/txpool` |
| [Transact](#transact-and-hold) | burrow.transactAndHold | POST | `/unsafe/txpool?hold=true` |
| [TransactNameReg](#transact-name-reg) | burrow.transactNameReg | POST | `/unsafe/namereg/txpool` |
| [GenPrivAccount](#gen-priv-account) | burrow.genPrivAccount | POST | `/unsafe/pa_generator` |

Here are the catagories.

* [Accounts](#accounts)
* [BlockChain](#blockchain)
* [Consensus](#consensus)
* [Events](#events)
* [Name-registry](#name-registry)
* [Network](#network)
* [Transactions](#transactions)
* [Code Execution (calls)](#calls)
* [Unsafe](#unsafe)

In the case of **JSON-RPC**, the parameters are wrapped in a request object, and the return value is wrapped in a response object.

In the case of **REST-like HTTP** GET requests, the params (and query) are provided in the url. If it's a POST, PATCH or PUT request, the parameter object should be written to the body of the request as JSON. It is normally the same params object as in JSON-RPC.

**Unsafe** are methods that require a private key to be sent either to or from the client and should therefore be used only during development/testing, or with extreme care. They may be phased out entirely.

<a name="accounts"></a>
### Accounts

***

<a name="get-accounts"></a>
#### GetAccounts

Get accounts will return a list of accounts. If no filtering is used, it will return all existing accounts.

##### HTTP

Method: GET

Endpoint: `/accounts`

##### JSON-RPC

Method: `burrow.getAccounts`

Parameter:

```
{
	filters: [<FilterData>]
}
```

##### Filters

| Field | Underlying type | Ops | Example Queries |
| :---- | :-------------- | :-- | :-------------- |
| `balance` | uint64 | `<`, `>`, `<=`, `>=`, `==` | `q=balance:<=11` |
| `code` | byte[] | `==`, `!=` | `q=code:1FA872` |

##### Return value

```
{
	accounts: [<Account>]
}
```

##### Additional info

See GetAccount below for more info on the `Account` object.

See the section on [Filters](#queries-filters) for info on the `FilterData` object.

***

<a name="get-account"></a>
#### GetAccount

Get an account by its address.

##### HTTP

Method: GET

Endpoint: `/accounts/:address`

Params: The public `address` as a hex string.


##### JSON-RPC

Method: `burrow.getAccount`

Parameter:

```
{
	address: <string>
}
```

##### Return value

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

##### Additional info

There are two types of objects used to represent accounts: public accounts (like the one here) and private accounts that only hold information about an account's address, and public and private keys.

***

<a name="get-storage"></a>
#### GetStorage

Get the complete storage of a contract account. Non-contract accounts have no storage.

NOTE: This is mainly used for debugging. In most cases the storage of an account would be accessed via public accessor functions defined in the contracts ABI.

##### HTTP

Method: GET

Endpoint: `/accounts/:address/storage`

Params: The public `address` as a hex string.


##### JSON-RPC

Method: `burrow.getStorage`

Parameter:

```
{
	address: <string>
}
```

##### Return value

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
#### GetStorageAt

Get a particular entry in the storage of a contract account. Non-contract accounts have no storage.

NOTE: This is mainly used for debugging. In most cases the storage of an account would be accessed via public accessor functions defined in the contracts ABI.

##### HTTP

Method: GET

Endpoint: `/accounts/:address/storage/:key`

Params: The public `address` as a hex string, and the `key` as a hex string.

##### JSON-RPC

Method: `burrow.getStorageAt`

Parameter:

```
{
	address: <string>
	key:     <string>
}
```

##### Return value

```
{
	key:   <string>
	value: <string>
}
```

Both `key` and `value` are hex strings.

***

<a name="blockchain"></a>
### Blockchain

***

<a name="get-blockchain-info"></a>
#### GetBlockchainInfo

Get the current state of the blockchain. This includes things like chain-id and latest block height. There are individual getters for all fields as well.

##### HTTP

Method: GET

Endpoint: `/blockchain`

##### JSON-RPC

Method: `burrow.getBlockchainInfo`

Parameter: -

##### Return value

```
{
	chain_id:            <string>
	genesis_hash:        <string>
	latest_block:        <BlockMeta>
	latest_block_height: <number>
}
```

##### Additional info

`chain_id` is the name of the chain.
`genesis_hash` is a 32 byte hex-string. It is the hash of the genesis block, which is the first block on the chain.
`latest_block` contains block metadata for the latest block. See the [GetBlock](#get-block) method for more info.
`latest_block_height` is the height of the latest block, and thus also the height of the entire chain.

The block *height* is sometimes referred to as the block *number*.

See [GetBlock](#get-block) for more info on the `BlockMeta` type.

***

<a name="get-chain-id"></a>
#### GetChainId

Get the chain id.

##### HTTP

Method: GET

Endpoint: `/blockchain/chain_id`

##### JSON-RPC

Method: `burrow.getChainId`

Parameter: -

##### Return value

```
{
	chain_id:            <string>
}
```
***

<a name="get-genesis-hash"></a>
#### GetGenesisHash

Get the genesis hash. This is a 32 byte hex-string representation of the hash of the genesis block. The genesis block is the first block on the chain.

##### HTTP

Method: GET

Endpoint: `/blockchain/genesis_hash`

##### JSON-RPC

Method: `burrow.getGenesisHash`

Parameter: -

##### Return value

```
{
	genesis_hash:        <string>
}
```

***

<a name="get-latest-block-height"></a>
#### GetLatestBlockHeight

Get the height of the latest block. This would also be the height of the entire chain.

##### HTTP

Method: GET

Endpoint: `/blockchain/latest_block/height`

##### JSON-RPC

Method: `burrow.getLatestBlockHeight`

Parameter: -

##### Return value

```
{
	latest_block_height: <number>
}
```

***

<a name="get-latest-block"></a>
#### GetLatestBlock

Gets the block that was added to the chain most recently.

##### HTTP

Method: GET

Endpoint: `/blockchain/latest_block`

##### JSON-RPC

Method: `burrow.getLatestBlock`

Parameter: -

##### Return value

```
{
	latest_block:        <BlockMeta>
}
```

##### Additional info

See [GetBlock](#get-block) for more info on the `BlockMeta` type.

***

<a name="get-blocks"></a>
#### GetBlocks

Get a series of blocks from the chain.

##### HTTP

Method: GET

Endpoint: `/blockchain/blocks`

##### JSON-RPC

Method: `burrow.getBlocks`

Parameter:

```
{
	filters: [<FilterData>]
}
```

##### Filters

| Field | Underlying type | Ops | Example Queries |
| :---- | :-------------- | :-- | :-------------- |
| `height` | uint | `<`, `>`, `<=`, `>=`, `==` | `q=height:>4`, `q=height:10..*` |



##### Return value

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

##### Additional info

TODO

See the section on [Filters](#queries-filters) for info on the `FilterData` object.

`min_height` and `max_height` are the two actual values used for min and max height when fetching the blocks. The reason they are included is because the heights might have been modified, like for example when the blockchain height is lower than the max height provided in the query.

See [GetBlock](#get-block) for more info on the `BlockMeta` type.

***

<a name="get-block"></a>
#### GetBlock

Get the block at the given height.

##### HTTP

Method: GET

Endpoint: `/blockchain/block/:number`

##### JSON-RPC

Method: `burrow.getBlock`

Parameter:

```
{
	height: <number>
}
```

##### Return value

```
{

	header: {
		chain_id:        <string>
		height:          <number>
		time:            <string>
		num_txs:         <number>
		last_block_id: {
			hash:	<string>
			parts: {
				total:	<int>
				hash:	<string>
			}
		}
		last_commit_hash:	<string>
		data_hash:			<string>
		validators_hast:	<string>
		app_hash:			<string>
	}
	data: {
		txs: [<Tx>]
	}
	last_commit: {
		blockID: {
			hash:	<string>
			parts: {
				total:	<int>
				hash:	<string>
			}
		}
		precommits: {
			validator_address:	<string>
			validator_index:	<int>
			height:				<int>
			round:				<int>
			type:				<int>
			block_id: {
				hash:	<string>
				parts: {
					total:	<int>
					hash:	<string>
				}
			}
			signature: [<signature>]
		}
	}
	id:		 <string>
	jsonrpc: <string>
}
```

The `Signature` object:

```
{
	index:		<int>
	signature:	<string>
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

##### Additional info

TODO

See [The transaction types](#the-transaction-types) for more info on the `Tx` types.

***

<a name="consensus"></a>
### Consensus

***

<a name="get-consensus-state"></a>
#### GetConsensusState

Get the current consensus state.

##### HTTP

Method: GET

Endpoint: `/consensus`

##### JSON-RPC

Method: `burrow.getConsensusState`

Parameter: -

##### Return value

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

##### Additional info

TODO

See the GetValidators method right below for info about the `Validator` object.

***

<a name="get-validators"></a>
#### GetValidators

Get the validators.

##### HTTP

Method: GET

Endpoint: `/consensus/validators`

##### JSON-RPC

Method: `burrow.getValidators`

Parameter: -

##### Return value

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

##### Additional info

TODO

***

<a name="events"></a>
### Events

***

<a name="event-subscribe"></a>
#### EventSubscribe

Subscribe to a given type of event. The event is identified by the `event_id` (see [Event types](#event-types). The response provides a subscription identifier `sub_id` which tracks your client and can be used to [unsubscribe](#eventunsubscribe).

##### HTTP

Method: POST

Endpoint: `/event_subs/`

Body: See JSON-RPC parameter.

##### JSON-RPC

Method: `burrow.eventSubscribe`

Parameter:

```
{
	event_id: <string>
}
```

##### Return value

```
{
	sub_id: <string>
}
```

##### Additional info

For more information about events and the event system, see the [Event system](#event-system) section.

***

<a name="event-unsubscribe"></a>
#### EventUnsubscribe

Unsubscribe to an event by supplying the subscription identifier `sub_id` you obtained from a previous call to [subscribe](#eventsubscribe).

##### HTTP

Method: DELETE

Endpoint: `/event_subs/:id`

##### JSON-RPC

Method: `burrow.eventUnsubscribe`

Parameter:
```
{
	sub_id: <string>
}
```

##### Return value

```
{
	result: <bool>
}
```

##### Additional info

For more information about events and the event system, see the [Event system](#event-system) section.

***

<a name="event-poll"></a>
#### EventPoll

Poll a subscription. Note, this cannot be done if using websockets because the events will be passed automatically over the socket.

##### HTTP

Method: GET

Endpoint: `/event_subs/:id`

##### JSON-RPC

Method: `burrow.eventPoll`

##### Return value

```
{
	events: [<Event>]
}
```

##### Additional info

For more information about events and the event system, see the [Event system](#event-system) section. This includes info about the `Event` object.

***


<a name="name-registry"></a>
#### Name-registry

<a name="get-namereg-entries"></a>
#### GetNameRegEntries

This will return a list of name reg entries. Filters may be used.

##### HTTP

Method: GET

Endpoint: `/namereg`

##### JSON-RPC

Method: `burrow.getNameRegEntries`

Parameter:

```
{
	filters: [<FilterData>]
}
```

##### Filters

| Field | Underlying type | Ops | Example Queries |
| :---- | :-------------- | :-- | :-------------- |
| `expires` | int | `<`, `>`, `<=`, `>=`, `==` | `q=expires:<=50` |
| `owner` | byte[] | `==`, `!=` | `q=owner:1010101010101010101010101010101010101010` |
| `name` | string | `==`, `!=` | `q=name:!=somekey` |
| `data` | string | `==`, `!=` | `q=name:!=somedata` |

NOTE: While it is supported, there is no point in using `name:==...`, as it would search the entire map of names for that entry. Instead you should use the method `GetNameRegEntry` which takes the name (key) as argument.

##### Return value

```
{
	block_height: <number>
	names:        <NameRegEntry>
}
```

##### Additional info

See GetNameRegEntry below for more info on the `NameRegEntry` object.

See the section on [Filters](#queries-filters) for info on the `FilterData` object.

***

<a name="get-namereg-entry"></a>
#### GetNameRegEntry

Get a namereg entry by its key.

##### HTTP

Method: GET

Endpoint: `/namereg/:name`

Params: The key (a string)


##### JSON-RPC

Method: `burrow.getNameRegEntry`

Parameter:

```
{
	name: <string>
}
```

##### Return value

```
{
	owner:   <string>
	name:    <string>
	data:    <string>
	expires: <number>
}
```

***

<a name="network"></a>
### Network

***

<a name="get-network-info"></a>
#### GetNetworkInfo

Get the network information. This includes the blockchain client moniker, peer data, and other things.

##### HTTP

Method: GET

Endpoint: `/network`

##### JSON-RPC

Method: `burrow.getNetworkInfo`

Parameters: -

##### Return value

```
{
	client_version: <string>
	moniker: <string>
	listening: <boolean>
	listeners: [<string>]
	peers: [<Peer>]
}
```

##### Additional info

`client_version` is the version of the running client, or node.
`moniker` is a moniker for the node.
`listening` is a check if the node is listening for connections.
`listeners` is a list of active listeners.
`peers` is a list of peers.

See [GetPeer](#get-peer) for info on the `Peer` object.

***

<a name="get-client-version"></a>
#### GetClientVersion

Get the version of the running client (node).

##### HTTP

Method: GET

Endpoint: `/network/client_version`

##### JSON-RPC

Method: `burrow.getClientVersion`

Parameters: -

##### Return value

```
{
	client_version: <string>
}
```

***

<a name="get-moniker"></a>
#### GetMoniker

Get the node moniker, or nickname.

##### HTTP

Method: GET

Endpoint: `/network/moniker`

##### JSON-RPC

Method: `burrow.getMoniker`

Parameters: -

##### Return value

```
{
	moniker: <string>
}
```

***

<a name="is-listening"></a>
#### IsListening

Check whether or not the node is listening for connections.

##### HTTP

Method: GET

Endpoint: `/network/listening`

##### JSON-RPC

Method: `burrow.isListening`

Parameters: -

##### Return value

```
{
	listening: <boolean>
}
```

***

<a name="get-listeners"></a>
#### GetListeners

Get a list of all active listeners.

##### HTTP

Method: GET

Endpoint: `/network/listeners`

##### JSON-RPC

Method: `burrow.getListeners`

Parameters: -

##### Return value

```
{
	listeners: [<string>]
}
```

***

<a name="get-peers"></a>
#### GetPeers

Get a list of all peers.

##### HTTP

Method: GET

Endpoint: `/network/peers`

##### JSON-RPC

Method: `burrow.getPeers`

Parameters: -

##### Return value

```
{
	peers: [<Peer>]
}
```

See [GetPeer](#get-peer) below for info on the `Peer` object.

***

<a name="get-peer"></a>
#### GetPeer

Get the peer with the given IP address.

##### HTTP

Method: GET

Endpoint: `/network/peer/:address`

##### JSON-RPC

Method: `burrow.getPeer`

Parameters:

```
{
	address: <string>
}
```

##### Return value

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


##### Additional info

TODO

***

<a name="transactions"></a>
### Transactions

***

<a name="broadcast-tx"></a>
#### BroadcastTx

Broadcast a given (signed) transaction to the node. It will be added to the tx pool if there are no issues, and if it is accepted by all validators it will eventually be committed to a block.

WARNING: BroadcastTx will not be useful until we add a client-side signing solution.

##### HTTP

Method: POST

Endpoint: `/txpool`

Body:

```
<Tx>
```

##### JSON-RPC

Method: `burrow.BroadcastTx`

Parameters:

```
<Tx>
```

##### Return value

```
{
	tx_hash:          <string>
	creates_contract: <number>
	contract_addr:    <string>
}
```

##### Additional info

`tx_hash` is the hash of the transaction (think digest), and can be used to reference it.

`creates_contract` is set to `1` if a contract was created, otherwise it is 0.

If a contract was created, then `contract_addr` will contain the address. NOTE: This is no guarantee that the contract will actually be commited to the chain. This response is returned upon broadcasting, not when the transaction has been committed to a block.

See [The transaction types](#the-transaction-types) for more info on the `Tx` types.

***

<a name="get-unconfirmed-txs"></a>
#### GetUnconfirmedTxs

Get a list of transactions currently residing in the transaction pool. These have been admitted to the pool but have not yet been committed.

##### HTTP

Method: GET

Endpoint: `/txpool`

##### JSON-RPC

Method: `burrow.getUnconfirmedTxs`

Parameters: -

##### Return value

```
{
	txs: [<Tx>]
}
```


##### Additional info

See [The transaction types](#the-transaction-types) for more info on the `Tx` types.

***

<a name="calls"></a>
### Code execution (calls)

***

<a name="call"></a>
#### Call

Call a given (contract) account to execute its code with the given in-data.

##### HTTP

Method: POST

Endpoint: `/calls`

Body: See JSON-RPC parameter.

##### JSON-RPC

Method: `burrow.call`

Parameters:

```
{
	address: <string>
	data: <string>
}
```

##### Return value

```
{
	return:   <string>
	gas_used: <number>
}
```

##### Additional info

`data` is a string of data formatted in accordance with the [contract ABI](https://github.com/monax/legacy-contracts.js)

***

<a name="call-code"></a>
#### CallCode

Pass contract code and tx data to the node and have it executed in the virtual machine. This is mostly a dev feature.

##### HTTP

Method: POST

Endpoint: `/codecalls`

Body: See JSON-RPC parameter.

##### JSON-RPC

Method: `burrow.callCode`

Parameters:

```
{
	code: <string>
	data: <string>
}
```

##### Return value

```
{
	return: <string>
	gas_used: <number>
}
```

##### Additional info

`code` is a hex-string representation of compiled contract code.
`data` is a string of data formatted in accordance with the [contract ABI](https://github.com/monax/legacy-contracts.js)

***

<a name="unsafe"></a>
### Unsafe

These methods are unsafe because they require that a private key is either transmitted or received. They are supposed to be used only in development.

***

<a name="transact"></a>
#### Transact

Convenience method for sending a transaction. It will do the following things:

* Use the private key to create a private account object (i.e. generate public key and address).
* Use the other parameters to create a `CallTx` object.
* Sign the transaction.
* Broadcast the transaction.

##### HTTP

Method: POST

Endpoint: `/unsafe/txpool`

Body: See JSON-RPC parameters.

##### JSON-RPC

Method: `burrow.transact`

Parameters:

```
{
	priv_key:  <string>
	data:      <string>
	address:   <string>
	fee:       <number>
	gas_limit: <number>
}
```

private key is the hex string only.

##### Return value

The same as with BroadcastTx:

```
{
	tx_hash:          <string>
	creates_contract: <number>
	contract_addr:    <string>
}
```

##### Additional info

See [The transaction types](#the-transaction-types) for more info on the `CallTx` type.

If you want to hold the tx, use `/unsafe/txpool?hold=true`. See `TransactAndHold` below.

***

<a name="transact-and-hold"></a>
#### TransactAndHold

Convenience method for sending a transaction and holding until it's been committed (or not). It will do the following things:

* Use the private key to create a private account object (i.e. generate public key and address).
* Use the other parameters to create a `CallTx` object.
* Sign the transaction.
* Broadcast the transaction.
* Wait until the transaction is fully processed.

When holding, the request will eventually timeout if the transaction is not processed or if it produces an error. The response will then be an error that includes the transaction hash (which can be used for further investigation).

##### HTTP

Method: POST

Endpoint: `/unsafe/txpool`

Query: `?hold=true`

Body: See JSON-RPC parameters.

##### JSON-RPC

Method: `burrow.transactAndHold`

Parameters:

```
{
	priv_key:  <string>
	data:      <string>
	address:   <string>
	fee:       <number>
	gas_limit: <number>
}
```

private key is the hex string only.

##### Return value

```
{
	call_data: {
		caller: <string>
    	callee: <string>
    	data:   <string>
    	value:  <number>
    	gas:    <number>
	}
	origin:     <string>
	tx_id:      <string>
	return:     <string>
	exception:  <string>
}
```

##### Additional info

See [The transaction types](#the-transaction-types) for more info on the `CallTx` type.

If you don't want to hold the tx, either use `/unsafe/txpool?hold=false` or omit the query entirely. See `Transact` for the regular version.

***

<a name="transact-name-reg"></a>
#### TransactNameReg

Convenience method for sending a transaction to the name registry. It will do the following things:

* Use the private key to create a private account object (i.e. generate public key and address).
* Use the other parameters to create a `NameTx` object.
* Sign the transaction.
* Broadcast the transaction.

##### HTTP

Method: POST

Endpoint: `/unsafe/namereg/txpool`

Body: See JSON-RPC parameters.

##### JSON-RPC

Method: `burrow.transactNameReg`

Parameters:

```
{
	priv_key:  <string>
	name:      <string>
	data:      <string>
	fee:       <number>
	amount:    <number>
}
```

##### Return value

The same as with BroadcastTx:

```
{
	tx_hash:          <string>
	creates_contract: <number> (always 0)
	contract_addr:    <string> (always empty)
}
```

##### Additional info

See [The transaction types](#the-transaction-types) for more info on the `NameTx` type.

***

<a name="gen-priv-account"></a>
#### GenPrivAccount

Convenience method for generating a `PrivAccount` object, which contains a private key and the corresponding public key and address.

##### HTTP

Method: POST

Endpoint: `/unsafe/pa_generator`
##### JSON-RPC

Method: `burrow.genPrivAccount`

Parameters: -

##### Return value

```
{
	address: <string>
	pub_key: <PubKey>
	priv_key: <PrivKey>
}
```

##### Additional info

TODO fix endpoint and method.

Again - This is unsafe. Be warned.

***

<a name="queries-filters"></a>
## Filters

Filters are used in searches. The structure is similar to that of the [Github api (v3)](https://developer.github.com/v3/search/).

### JSON-RPC

Filters are added as objects in the request parameter. Methods that support filtering include an array of filters somewhere in their params object.

Filter:

```
{
    field: <string>
    op:    <string>
    value: <string>
}
```

* The `field` must be one that is supported by the collection items in question.
* The `op` is a relational operation `[>, <, >=, <=, ==, !=]`. Different fields supports different subsets.
* The `value` is the value to match against. It is always a string.
* Range queries are done simply by adding two filters - one for the minimum value and one for the maximum.

##### Examples

We want an account filter that only includes accounts that have code in them (i.e. contract accounts):

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

### HTTP Queries

The structure of a normal query is: `q=field:[op]value+field2:[op2]value2+ ... `.

- `q` means it's a query.
- `+` is the filter separator.
- `field` is the field name.
- `:` is the field:relation separator.
- `op` is the relational operator, `>, <, >=, <=, ==, !=`.
- `value` is the value to compare against, e.g. `balance:>=5` or `language:==golang`.

There is also support for [range queries](https://help.github.com/articles/search-syntax/): `A..B`, where `A` and `B` are number-strings. You may use the wildcard `*` instead of a number. The wildcard is context-sensitive; if it is put on the left-hand side, it is the minimum value, and, if on the right-hand side, it is the maximum value. Let `height` be an unsigned byte with no additional restrictions. `height:*..55` would then be the same as `height:0..55`, and `height:*..*` would be the same as `height:0..255`.

NOTE: URL encoding applies as usual. Omitting it here for clarity.

`op` will default to (`==`) if left out, meaning `balance:5` is the same as `balance:==5`

`value` may be left out if the field accepts the empty string as input. This means if `code` is a supported string type,  `code:==` would check if the code field is empty. We could also use the inferred `==` meaning this would be equivalent: `code:`.  The system may be extended so that the empty string is automatically converted to the null-type of the underlying field, no matter what that type is. If balance is a number, then `balance:` would be the same as `balance:==0` (and `balance:0`).

##### Example

We want to use the same filter as in the JSON version; one that finds all contract accounts.

`q=code:!=`

One that finds those with balance less then 1000:

`q=balance:<1000`

One that finds those where 0 <= balance <= 1000.

`q=balance:0..1000`

One that finds non-contract accounts with 0 <= balance <= 1000:

`q=balance:0..1000+code:`
