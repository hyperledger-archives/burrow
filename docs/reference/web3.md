# Web3 JSON RPC

Burrow now ships with a web3 compliant RPC server to integrate with your favorite Ethereum tooling!
We've already tried a few tools to ensure they work correctly, but if you have any problems please 
consider submitting a pull request.

## Blockscout

[Blockscout](https://github.com/poanetwork/blockscout) is a graphical blockchain explorer for 
Ethereum based networks. Before deploying the application, ensure to set the following environment 
variables so it can locate your local Burrow node.

```bash
export ETHEREUM_JSONRPC_VARIANT=ganache
export ETHEREUM_JSONRPC_HTTP_URL=http://localhost:26660
```

## Metamask

[Metamask](https://metamask.io/) is an open-source identity management application for Ethereum, 
typically used as a browser extension. After creating or importing a supported `secp256k1` key pair, 
you can simply add Burrow to the list of networks.

## Remix

[Remix](https://remix.ethereum.org/) is a web-based integrated development environment for Solidity.
To deploy and run transactions, select `Web3 
Provider` as the `Environment` and enter your local RPC
address when prompted.

## Truffle

[Truffle](https://www.trufflesuite.com/docs/truffle/overview) makes it easy to develop smart contracts 
with automatic compilation, linking and deployment. For a quick introduction, follow the official 
[tutorial](https://www.trufflesuite.com/docs/truffle/quickstart) and edit the config `truffle-config.js` 
to point to your local node. To ensure Truffle uses this configuration, simply suffix all commands with 
the flag `--network burrow`. 

```js
module.exports = {
  networks: {
   burrow: {
     host: "127.0.0.1",
     port: 26660,
     network_id: "*"
   },
  }
};
```

