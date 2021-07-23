# JS API

Burrow has a JavaScript API for communicating with a [Hyperledger Burrow](https://github.com/hyperledger/burrow) server, which implements the GRPC spec.

## Prerequisites

- Burrow version 0.20 or higher
- Node.js version 7 or higher

You can check the installed version of Node.js with the command:

```bash
$ node --version
```

If your distribution of Linux has a version older than 6 then you can update it.

## Install

``` bash
$ yarn install @hyperledger/burrow
```

## Usage

You will need to know the <IP Address>:<PORT> of the burrow instance you wish to connect to. If running locally this will be 'localhost' and the default port, which is '10997'. DO NOT INCLUDE A PROTOCOL.

The main class is `Burrow`. A standard `Burrow` instance is created like this:

```JavaScript
const monax = require('@monax/burrow');
var burrowURL = "<IP address>:<PORT>"; // localhost:10997 if running locally on default port
var account = 'ABCDEF01234567890123'; // address of the account to use for signing, hex string representation 
var options = {objectReturn: true};
var burrow = monax.createInstance(burrowURL, account, options);
```

The parameters for `createInstance` is the server URL as a string or as an object `{host:<IP Address>, port:<PORT>}`. An account in the form of a hex-encoded address must be provided. 

> Note: the instance of burrow you are connecting to must have the associated key (if you want local signing you should be running a local node of burrow. Other local signing options might be made available at a later point). 

And finally an optional options object. Allowed options are:

* objectReturn: If True, communicating with contracts an object returns an object of the form: `{values:{...}, raw:[]}` where the values objects attempts to name the returns based on the abi and the raw is the decoded array of return values. If False just the array of decoded return values is returned.


## API Reference

There are bindings for all the GRPC methods. All functions are on the form `function(param1, param2, ... [, callback])`, where the callback is a function on the form `function(error, data)`. The `data` object is the same as you would get by calling the corresponding RPC method directly. If no callback is provided, a promise will be returned instead. If calling a response streaming GRPC call, the callback is not optional and will be called with `data` anytime it is recieved.

The structure of the library is such that there are lower-level access to the GRPC services and higher level wrappers of these services. The structure of the library is outlined below

### Burrow

The table below links to the reference schema for either the protobuf files governing the component or the Javascript interface.

| Component Name | Accessor |
| :----------- | :--------------- |
| Transactions | [Burrow.transact](https://github.com/hyperledger/burrow/blob/main/protobuf/rpctransact.proto) |
| Queries | [Burrow.query](https://github.com/hyperledger/burrow/blob/main/protobuf/rpcquery.proto) |
| EventStream | [Burrow.eventStream](https://github.com/hyperledger/burrow/blob/main/protobuf/rpcevents.proto) |
| Events | [Burrow.events](https://github.com/hyperledger/burrow/blob/main/lib/events.js) |
| NameReg | [Burrow.namereg](https://github.com/hyperledger/burrow/blob/main/lib/namereg.js) |

| Contracts | [Burrow.contracts](https://github.com/hyperledger/burrow/blob/main/lib/contractManager.js) |

### GRPC Access Components

Burrow provides access to three GRPC services; Transactions, Queries, and ExecutionEvents in the form of automatically generated code endpoints. Below details how to access these methods and links to the request and return objects defined in the protobuf specs. The format for all calls is `function(object[, callback])` The callback is optional for non-streaming endpoints in which case a promise will be returned. 

#### Sending Funds and Creating Accounts

In Burrow an account must already exist in order to be used as a input account (the sender of a transaction). An account can be created once the chain is running (accounts at genesis can be described in the genesis document in the accounts section) in the following ways:

1. Issuing a `SendTx` to the address to be created (see below) - where the input account must have both the `Send` and `CreateAccount` permission.
2. Sending value to the address to created from a smart contract using the CALL opcode - where the caller contract must have the `CreateAccount` permission.
3. Issuing a `GovTx` where accounts can be created/updated in bulk provided the input has the `Root` permission.

The conventional way to create an new account to use as an input for transactions is the following:

First create a key - you will want to create an account for which you have access to the private key controlling that account (as defined by the address of the public key):

```shell
# Create a new key against the key store of a locally running Burrow (or burrow keys standalone server):
$ address=$(burrow keys gen -n --name NewKey)

# The address will be printed to stdout so the above captures it in $address, you can also list named keys:
$ burrow keys list
Address:"6075EADD0C7A33EE6153F3FA1B21E4D80045FCE2" KeyName:"NewKey"
```

Note creating the key _does not_ create the account - it just generates a key pair in your local key store (it is not in anyway known the blockchain network).

Now we would like to use a `SendTx` to create the address, here's how to do that in Javscript:

```javascript
// Using account and burrow defined in snippet from [Usage](#usage)

// Address we want to create
var addressToCreate = "6075EADD0C7A33EE6153F3FA1B21E4D80045FCE2"

// The amount we send is arbitrary
var amount = 20

client.transact.SendTxSync(
  {
    Inputs: [{
      Address: Buffer.from(account, 'hex'),
      Amount: amount
    }],
    Outputs: [{
      Address: Buffer.from(addressToCreate, 'hex'),
      Amount: amount
    }]
  })
  .then(txe => console.log(txe))
  .catch(err => console.error(err))
```

The return `txe` (short for `TxExecution`) logged to the console in the `then` block  contains the history and fingerprint of the `SendTx` execution. You can see an example of this in [basic app](example/basic-app/app.js). 

#### NameReg access

Here is an example of usage in setting and getting a name:

```javascript
var setPayload = {
  Input: {
    Address: Buffer.from(account, 'hex'),
    Amount: 50000
  },
  Name: "DOUG",
  Data: "Marmot",
  Fee: 5000
}

var getPayload = {Name: "DOUG"}

// Using a callback
client.transact.NameTxSync(setPayload, function (error, data) {
  if (error) throw error; // or something more sensible
  // data object contains detailed information of the transaction execution.

  // Get a name this time using a promise
  client.query.GetName(getPayload)
    .then((data) => {
      console.log(data);
    }) // should print "Marmot"
    .catch((error) => {
      throw error;
    })
})

```

#### Transactions

`burrow.transact` provides access to the burrow GRPC service `rpctransact`. As a GRPC wrapper all the endpoints take a data argument and an optional callback. The format of the data object is specified in the [rpctransact protobuf file](./protobuf/rpctransact.proto).  A note on RPC naming, any method which ends in `Sync` will wait until the transaction generated is included in a block. Any `Async` method will return a receipt of the transaction immediately but does not guarantee it has been included. `Sim` methods request that the transaction be simulated and the result returned as if it had been executed. SIMULATED CALLS DO NOT GET COMMITTED AND DO NOT CHANGE STATE.

| Method | Passed | Returns |
| :----- | :--------- | :---- |
| burrow.transact.BroadcastTxSync | [TxEnvelopeParam](https://github.com/hyperledger/burrow/blob/main/protobuf/rpctransact.proto#L74-L79) | [TxExecution](https://github.com/hyperledger/burrow/blob/develop/protobuf/exec.proto#L34-L56) |
| burrow.transact.BroadcastTxASync | [TxEnvelopeParam](https://github.com/hyperledger/burrow/blob/main/protobuf/rpctransact.proto#L74-L79) | [Receipt](https://github.com/hyperledger/burrow/blob/develop/protobuf/txs.proto#L38-L47) |
| burrow.transact.SignTx | [TxEnvelopeParam](https://github.com/hyperledger/burrow/blob/main/protobuf/rpctransact.proto#L74-L79) | [TxEnvelope](https://github.com/hyperledger/burrow/blob/develop/protobuf/rpctransact.proto#L70-L72) |
| burrow.transact.FormulateTx | [PayloadParam](https://github.com/hyperledger/burrow/blob/main/protobuf/rpctransact.proto#L64-L68) | [TxEnvelope](https://github.com/hyperledger/burrow/blob/develop/protobuf/rpctransact.proto#L70-L72) |
| burrow.transact.CallTxSync | [CallTx](https://github.com/hyperledger/burrow/blob/main/protobuf/payload.proto#L53-L66) | [TxExecution](https://github.com/hyperledger/burrow/blob/develop/protobuf/exec.proto#L34-L56) |
| burrow.transact.CallTxAsync | [CallTx](https://github.com/hyperledger/burrow/blob/main/protobuf/payload.proto#L53-L66) | [Receipt](https://github.com/hyperledger/burrow/blob/develop/protobuf/txs.proto#L38-L47) |
| burrow.transact.CallTxSim | [CallTx](https://github.com/hyperledger/burrow/blob/main/protobuf/payload.proto#L53-L66) | [TxExecution](https://github.com/hyperledger/burrow/blob/develop/protobuf/exec.proto#L34-L56) |
| burrow.transact.SendTxSync | [SendTx](https://github.com/hyperledger/burrow/blob/main/protobuf/payload.proto#L69-L76) | [TxExecution](https://github.com/hyperledger/burrow/blob/develop/protobuf/exec.proto#L34-L56) |
| burrow.transact.SendTxAsync | [SendTx](https://github.com/hyperledger/burrow/blob/main/protobuf/payload.proto#L69-L76) | [Receipt](https://github.com/hyperledger/burrow/blob/develop/protobuf/txs.proto#L38-L47) |
| burrow.transact.NameTxSync | [NameTx](https://github.com/hyperledger/burrow/blob/main/protobuf/payload.proto#L88-L98) | [TxExecution](https://github.com/hyperledger/burrow/blob/develop/protobuf/exec.proto#L34-L56) |
| burrow.transact.NameTxAsync | [NameTx](https://github.com/hyperledger/burrow/blob/main/protobuf/payload.proto#L88-L98) | [Receipt](https://github.com/hyperledger/burrow/blob/develop/protobuf/txs.proto#L38-L47) |


#### Queries

`Burrow.query` provides access to the burrow GRPC service `rpcquery`. As a GRPC wrapper all the endpoints take a data argument and an optional callback. The format of the data object is specified in the [protobuf files](https://github.com/hyperledger/burrow/tree/main/js/protobuf). Note that "STREAM" functions take a callback `function(error, data)` which is mandatory and is called any time data is returned. For list Accounts the queryable tags are Address, PublicKey, Sequence, Balance, Code, Permissions (Case sensitive). As an example you can get all accounts with a balance greater than 1000 by `burrow.query.ListAccounts('Balance > 1000', callback)`. Multiple tag criteria can be combined using 'AND' and 'OR' for an example of a combined query see [here](https://github.com/hyperledger/burrow/blob/develop/protobuf/rpcevents.proto#L87). Similarly for ListNames, the avaible tags are Name, Data, Owner and Exires (once again case sensitive) use is identical to List accounts.

| Method | Passed | Returns | Notes |
| :----- | :--------- | :---- | :------- |
| burrow.query.GetAccount | [GetAccountParam](https://github.com/hyperledger/burrow/blob/main/protobuf/rpcquery.proto#L25-L27) | [ConcreteAccount](https://github.com/hyperledger/burrow/blob/develop/protobuf/acm.proto#L23-L31) | |
| burrow.query.ListAccounts | [ListAccountsParam](https://github.com/hyperledger/burrow/blob/main/protobuf/rpcquery.proto#L29-L31) | [ConcreteAccount](https://github.com/hyperledger/burrow/blob/develop/protobuf/acm.proto#L23-L31) | STREAM |
| burrow.query.GetNameParam | [GetNameParam](https://github.com/hyperledger/burrow/blob/main/protobuf/rpcquery.proto#L33-L35) | [Entry](https://github.com/hyperledger/burrow/blob/develop/protobuf/names.proto#L22-L32) | |
| burrow.query.ListNames | [ListNamesParam](https://github.com/hyperledger/burrow/blob/main/protobuf/rpcquery.proto#L37-L39) | [Entry](https://github.com/hyperledger/burrow/blob/develop/protobuf/names.proto#L22-L32) | STREAM|

#### EventStream

NB: When listening to contract events it is easier to use the contract interface (described below)

`Burrow.executionEvents` provides access to the burrow GRPC service `ExecutionEvents`. As a GRPC wrapper all the endpoints take a data argument and an optional callback. The format of the data object is specified in the [protobuf files](https://github.com/hyperledger/burrow/tree/main/js/protobuf). Note that "STREAM" functions take a callback `function(error, data)` which is mandatory and is called any time data is returned.

| Method | Passed | Returns | Notes |
| :----- | :--------- | :---- | :------- |
| burrow.executionEvents.GetBlock | [GetBlockRequest](https://github.com/hyperledger/burrow/blob/main/protobuf/rpcevents.proto#L37-L42) | [BlockExecution](https://github.com/hyperledger/burrow/blob/develop/protobuf/exec.proto#L20-L27) | |
| burrow.executionEvents.GetBlocks | [BlocksRequest](https://github.com/hyperledger/burrow/blob/main/protobuf/rpcevents.proto#L51-L89) | [BlockExecution](https://github.com/hyperledger/burrow/blob/develop/protobuf/exec.proto#L20-L27) | STREAM |
| burrow.executionEvents.GetTx | [GetTxRequest](https://github.com/hyperledger/burrow/blob/main/protobuf/rpcevents.proto#L44-L49) | [TxExecution](https://github.com/hyperledger/burrow/blob/develop/protobuf/exec.proto#L34-L56) | |
| burrow.executionEvents.GetTxs | [BlocksRequest](https://github.com/hyperledger/burrow/blob/main/protobuf/rpcevents.proto#L51-L89) | [GetTxsResponse](https://github.com/hyperledger/burrow/blob/develop/protobuf/rpcevents.proto#L96-L99) | STREAM |
| burrow.executionEvents.GetEvents | [BlocksRequest](https://github.com/hyperledger/burrow/blob/main/protobuf/rpcevents.proto#L51-L89) | [GetEventsResponse](https://github.com/hyperledger/burrow/blob/develop/protobuf/rpcevents.proto#L91-L94) | STREAM |


***

### High-Level Components

In addition to direct access to the grpc services, the burrow object also provides access to *three* higher level components which wrap the low level access for convenience. These are ```.namereg```, ```.events```, and ```.contracts```. All high level components use the account provided during creation of the burrow instance for constructing transactions. Component ```contracts``` is the most important of the three, as events and namereg are really just helpful wrappers.


#### 1. Namereg

`burrow.namereg` is a convenience wrapper for setting and getting entries from the name registry. 
 
##### burrow.namereg.get
```burrow.namereg.get(name[,callback])```

Gets an entry stored at the name. It returns a promise if callback not provided.
###### Parameters
1. `String` - Name you wish to retrieve from the namereg
2. `function` - (optional) Function to call upon completion of form `function(error, data)`.
###### Returns
`Object` - The return data object is of the form:

```javascript
{
    Name: (registered name) (string)
    Owner: (address of name owner) (buffer)
    Data: (stored data) (string)
    Expires: (block at which entry expires) (int)
} 
```

##### burrow.namereg.set
```burrow.namereg.set(name, data, lease[, callback])```

Sets an entry in the namereg. It returns a promise if callback not provided.
###### Parameters
1. `String` - The name you wish to register
2. `String` - The string data you wish to store at the registered name (longer string = larger fee)
3. `int` - The number of blocks to register the name for (more blocks = larger fee)
4. `function` - (optional) Function to call upon completion of form `function(error, data)`.
###### Returns
`TxExecution` - The return data object is a [TxExecution](https://github.com/hyperledger/burrow/blob/main/protobuf/exec.proto#L34-L56).
###### Example

```javascript
// Using a callback
client.namereg.set("DOUG", "Marmot", 5000, function (error, data) {
  if (error) throw error; // or something more sensible
  // data object contains detailed information of the transaction execution.

  // Get a name this time using a promise
  client.namereg.get("DOUG")
    .then((data) => {
      console.log(data);
    }) // Should print "Marmot"
    .catch((error) => {
      throw error;
    })
})
```

> Note: this example is nearly identical to the example above except that the objects are not explicitly constructed by you.



#### 2. Events

`burrow.events` contains convenience wrappers for streaming executionEvents.

##### burrow.events.listen
```burrow.events.listen(query, options, callback)```

Listens to execution events which satisfy the filter query.
###### Parameters
1. `String` - a pegjs querystring for filtering the returned events see [here]() for grammar specification
2. `Object` - Currently unused. pass `{}`
3. `function` - Signature of `function(error, data)` mandatory
###### Returns
`GetEventsResponse` - The return data object is a [GetEventsResponse](https://github.com/hyperledger/burrow/blob/main/protobuf/rpcevents.proto#L91-L94)



##### burrow.events.subContractEvents
```burrow.events.subContractEvents(address, signature, options, callback)```

Listens to EVM event executions from specific contract.
###### Parameters
1. `String` - hex string of the contract address of interest
2. `String` - event abi signature
3. `Object` - Currently unused. pass `{}`
4. `function` - Signature of `function(error, data)` mandatory.
###### Returns
`GetEventsResponse` - The return data object is a [GetEventsResponse](https://github.com/hyperledger/burrow/blob/main/protobuf/rpcevents.proto#L91-L94)



#### 3. Contracts

`burrow.contracts` is arguably the most important component of the burrow it exposes two functions, `.deploy` and `.new` both of which return a Contract interface object (sometimes refered to as contract object). The difference between them is that `new` simply creates an interface to a contract and `deploy` will first create an instance and then deploy a copy of it to the blockchain.


##### burrow.contracts.deploy
```burrow.contracts.deploy(abi, bytecode, params... [, callback])```

Deploys a contract and returns a contract interface either to the callback or a promise once deploy is successful. It returns a promise if callback not provided.

When the contract interface object is created via deploy, the default address is set to the address of the deployed contract (which can be accessed as contract.address). This interface object can still be used as a generic interface but care must be taken to use the `.at()` and `.atSim()` versions of functions.

###### Parameters
1. `Object` - the object corresponding to the json ABI of the contract you wish to interface with.
2. `String` - Hex encoded string of bytecode of the contract to deploy
3. `params` - arguments to the constructor function (if there are any)
4. `function` - (optional) Format of `function(error, contract)` where contract is the contract interface object.
###### Returns
`Object` - The return data object is a contract interface, which refers to the contract which is deployed at `contract.address`. (This functionality used to be called `new`.)

##### burrow.contracts.new
```burrow.contracts.address(address)```

Returns a new contract interface object, without having to pass in the ABI. The ABI is retrieved from burrow; the contract must have been deployed with burrow deploy 0.28.0 or later.

###### Parameters
3. `String` - Hex encoded address of the default contract you want the interface to access
###### Returns
`Object` - The return data object is a contract interface.

##### burrow.contracts.new
```burrow.contracts.new(abi, [bytecode[, address]])```

Returns a new contract interface object. All you really need to create an interface is the abi, however you can also include the bytecode of the contract. If you do so you can create new contracts of this type by calling `contract._constructor(...)` which will deploy a new contract and return its address. If you provide an address, then this will be the default contract address used however you can also omit this at be sure to use the `.at()` and `.atSim()` versions of functions. Also note you must provide bytecode is you wish to provide address, though bytecode argument can be null.

###### Parameters
1. `Object` - the object corresponding to the json ABI of the contract you wish to interface with.
2. `String` - Hex encoded string of bytecode of the contract to deploy
3. `String` - (optional) Hex encoded address of the default contract you want the interface to access
###### Returns
`Object` - The return data object is a contract interface.


#### 3.1. Contract interface object

The contract interface object allows for easy access of solidity contract function calls and subscription to events. When created, javascript functions for all functions specified in the abi are generated. All of these functions have the same form `Contract.functionname(params...[, callback])`, where `params` are the arguments to the contract constructor. Arguments of the "bytes" type should be properly hex encoded before passing, to avoid improper encoding. If a callback is not provided a promise is returned.

> Note: if the burrow object was created with ```{objectReturn: True}``` the return from these function calls is formatted as `{values:{...}, raw:[]}` otherwise an array of decoded values is provided. The values object names the decoded values according to the abi spec, if a return value has no name it won't be included in the values object and must be retrieved from its position on the raw array.


In the case of a REVERT op-code being called in the contract function call, an error will be passed with the revert string as the `.message` field. These errors can be distinguished from other errors as the `.code` field will be `ERR_EXECUTION_REVERT`.

In addition to the standard function call, there are three other forms: `contract.functionname.sim`, `contract.functionname.at`, `contract.functionname.atSim`.


##### contract.functionname.sim
```contract.functionname.sim(params... [, callback])```

The "Sim" forms will force a simulated call so that does not change state. Although, the data returned is identical to what would have been returned if the call had been submitted. Useful for querying data or checking if a transaction passes some tests.
###### Parameters
1. `params` - the arguments to the function (if there are any)
2. `function`- (optional) Function to call upon completion of form `function(error, data)`.



##### contract.functionname.at
```contract.functionname.at(address, params... [, callback])```

The "at" forms allow you to specify which contract you wish to submit the transaction to. This allows you to use a single contract interface instance to access any contract with the same abi. Useful if for example there is a factory contract on the chain and you wish to connect to any of its children. The at forms MUST be used if a default address was not provided or created.
###### Parameters
1. `String` - Hex encoded address of the default contract you want the interface to access
2. `params` - the arguments to the function (if there are any)
3. `function`- (optional) Function to call upon completion of form `function(error, data)`.



##### contract.functionname.atSim
```contract.functionname.at(address, params... [, callback])```


###### Parameters
1. `String` - Hex encoded address of the default contract you want the interface to access
2. `params` - the arguments to the function (if there are any)
3. `function`- (optional) Function to call upon completion of form `function(error, data)`


##### contract._constructor
```contract._constructor(params... [, callback])```

Deploys a new contract from the same interface (no need to create a new interface object via .deploy). Once completed it will return the created contract's address.


###### Parameters
1. `params` - the arguments to the function (if there are any)
3. `function`- (optional) Function to call upon completion of form `function(error, data)`.
###### Returns
`String` - The return data String is the created contract's address.


#### 3.2. Encoding and Decoding Params

Occassionally you may wish to encode the parameters to a function call but not actually make the call. The most common use of this is in forwarding contracts which take pre-encoded arguments along with function signature bytes and then call another function passing that data for specifying the call. 

The Contract interface object supports this use case through `Contract.functionname.encode(...args)` which will return a hex string with the encoded arguments. This functionality is also available through `Monax.utils.encode(abi, functionname, ...args)`. In addition the complement also exists, `Contract.functionname.decode(data)` will produce the return object as if the data was just returned from a call.


#### 3.3. Contract Events

##### contract.eventname
```contract.eventname(callback)```

The contract interface object exposes subscription to Solidity events under event's name.
 where the provided callback with be passed an error and data of the form:

###### Parameters
1. `function` - Function to call upon completion of form `function(error, data)`. The data object has the following form:
```
{
	event: [fulleventname],
	address: [address of contract emitting event],
	args: {argname: argvalue}
}
```


##### contract.eventname.at
```contract.eventname.at(address, callback)```

Similarly to functions' contract it is possible to start listening to a non-default contract address. 

###### Parameters
1. `String` - hex string of the contract address of interest
2. `function` - Function to call upon completion of form `function(error, data)`. The data object has the following form:
```
{
	event: [fulleventname],
	address: [address of contract emitting event],
	args: {argname: argvalue}
}
```




### Example:

The following contract is a simple contract which takes a "name" to the constructor and also has a function `getName` which returns the name.

````solidity
pragma solidity ^0.4.18;
contract SimpleContract {

  string private name;

  function SimpleContract(string _newName) public {
    name = _newName;
  }

  function getName() public constant returns (string thename) {
    return name;
  }

}
````

Here I provide an example of communicating to the contract above from start to finish:

```javascript
const monax = require('@monax/burrow');
const assert = require('assert');

var burrowURL = "<IP address>:<PORT>"; // localhost:10997 if running locally on default port
var account = 'ABCDEF01234567890123'; // address of the account to use for signing, hex string representation 
var options = {objectReturn: false};
var burrow = monax.createInstance(burrowURL, account, options);

// Get the contractABIJSON from somewhere such as solc
var abi = json.parse(contractABIJSON) // Get the contractABIJSON from somewhere such as solc
var bytecode = contractBytecode // Get this from somewhere such as solc



// I'm going to use new to create a contract interface followed by a double direct call to the _constructor to deploy two contracts
const contract = burrow.contracts.new(abi, bytecode);
return Promise.all(								// Deployment of two contracts
	[
		contract._constructor('contract1'),
		contract._constructor('contract2')
	]
).then( ([address1, address2]) => {				// Collection of contracts' addresses
	console.log(address1 + " - contract1");
	console.log(address2 + " - contract2");
	return Promise.all(							// Execution of getName functions
		[
			contract.getName.at(address1),  // Using the .at() to specify the second deployed contract
			contract.getName.at(address2)
		]
	).then( ([name1, name2]) => {				// Collection of contracts' names
		console.log(address1 + " - " + name1);
		assert.equal(name1, 'contract1');
		console.log(address2 + " - " + name2);
		assert.equal(name2, 'contract2');
	});
});
```

<!-- 

## Documentation

Generate documentation using the command `yarn run doc`.

## Testing

To test the library against pre-recorded vectors:

```
yarn test
```

To test the library against Burrow while recording vectors:

```
TEST=record yarn test
```

To test Burrow against pre-recorded vectors without exercising the library:

```
TEST=server yarn test
```

## Debugging

Debugging information will display on `stderr` if the library is run with `NODE_DEBUG=monax` in the environment. -->

