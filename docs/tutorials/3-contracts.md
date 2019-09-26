# Contracts

Burrow supports both [Solidity](https://solidity.readthedocs.io/) and [WASM](reference/wasm.md) smart contracts. You may be familiar with this former language
if you have worked previously with Ethereum. If so, you will be pleased to know that Burrow can be used with [Remix](http://remix.ethereum.org/).

## Getting Started

Let's start a chain with a single validator:

```shell
burrow spec -v1 | burrow configure -s- | burrow start -c-
```

## Deploy Artifacts

For this step, we need two things: one or more solidity contracts and a deploy file. Let's take a simple example, found in [this directory](https://github.com/hyperledger/burrow/tree/master/tests/jobs_fixtures/app06-deploy_basic_contract_and_different_solc_types_packed_unpacked).

We need `deploy.yaml` and `storage.sol` in the same directory with **no other yaml or sol files**.

> [Solc](https://solidity.readthedocs.io/en/v0.4.21/installing-solidity.html) is required to compile Solidity code.

From inside that directory, we are ready to deploy.

```bash
burrow deploy --address $ADDRESS deploy.yaml
```

Replace `$ADDRESS` with the address at the top of your `burrow.toml`.

That's it! You've successfully deployed (and tested) a Solidity contract to a Burrow node.