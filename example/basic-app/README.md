# Example basic-app

This example contains an example solidity contract [simplestorage](simplestorage.sol) and a [node.js application](app.js) that interacts with the contract using [burrow module](../../js/README.md). It also contains a [makefile](makefile) that will set up a single node chain, deploy the contract using `burrow deploy`. The node app configures itself to use the the single node chain my looking for [account.json](account.json) and [deploy.output.json](deploy.output.json) files that are emitted by `burrow deploy` and the makefile.

The makefile provides some examples of using the `burrow` command line tooling and you are invited to modify it for your purposes (i.e. change from linux target to darwin)

## Dependencies
To run the makefile you will need to have installed:

- GNU Make
- Node.js (the `node` binary)
- npm (the node package manager)
- jq (the JSON tool)
- GO
- Solc (solidity compiler)

Burrow will be downloaded for you when using the makefile, but you may override `BURROW_BIN` and `BURROW_ARCH` in the makefile to change this behaviour. By default Burrow is downloaded for `Linux_x86_64.

## Running the example

All commands should be run from the same directory as this readme file.

### Step one
Start the chain

```shell
make start_chain
```

This will install burrow, create a new chain as necessary.

If successful you will see continuous output in your terminal, you can shutdown Burrow by sending the interrupt signal with Ctrl-C, and restart it again with whatever state has accumulated with `make start_chain`. If you would like to destroy the existing chain and start completely fresh (including deleting keys) run `make rechain`. If you would like to keep existing keys and chain config run `make reset_chain`.

You can redeploy the contract (to a new address) with `make redeploy`. The node app will then use this new contract by reading the address deploy.output.json. Be sure to do this if you wish to modify simplestorage.sol.

### Step two
Leave burrow running and in a separate terminal start the app which runs a simple HTTP server with:

```shell
make start_app
```

This will deploy the contract if necessary, install any node dependencies, and then start an expressjs server, which will run until interrupted.

### Step three
In a third terminal you may run the following commands that will call the Solidity smart contract using Javascript and burrow module:

```shell
# Inspect current value
  $ curl http://127.0.0.1:3000
  
# Set the value to 2000
  $ curl -d '{"value": 2000}' -H "Content-Type: application/json" -X POST http://127.0.0.1:3000
  
# Set the value via a testAndSet operation
  $ curl -d '{"value": 30}' -H "Content-Type: application/json" -X POST http://127.0.0.1:3000/2000
  
# Attempt the same testAndSet which now fails since the value stored is no longer '2000'
  $ curl -d '{"value": 30}' -H "Content-Type: application/json" -X POST http://127.0.0.1:3000/2000
  $ curl http://127.0.0.1:3000
```

Note: [httpie](https://httpie.org/) is a useful tool that makes POSTing to JSON endpoints more succinct:

```shell
# Inspect current value
  $ http 127.0.0.1:3000
  
# Set the value to 2000
  $ http 127.0.0.1:3000 value=2000
  
# Set the value via a testAndSet operation
  $ http 127.0.0.1:3000/2000 value=30
  
# Attempt the same testAndSet which now fails since the value stored is no longer '2000'
  $ http 127.0.0.1:3000/2000 value=30
  $ http 127.0.0.1:3000
```

## Additional endpoints

### Sending value and creating accounts

An endpoint is provided at `/send` for sending value from the default input account (defined in account.json after making the chain) that can be used with:

```shell
# Send 1000 units of native token to D1293FE65A071A9DB539D64858F0A3D4BCAA2EBA
  http :3000/send/D1293FE65A071A9DB539D64858F0A3D4BCAA2EBA amount=1000
```
