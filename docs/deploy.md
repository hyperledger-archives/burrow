# Burrow Deploy (Playbooks)

The Burrow deploy toolkit can do a number of things:

* compile Solidity source files (using solc) and deploy to chain
* call function on existing contract
* read or write to name registry
* manage permissions of accounts
* run tests and assert on result
* bond and unbond validators
* create proposals or vote for a proposal

burrow deploy needs a script to its commands. This script format bares some similarity to [ansible](https://www.ansible.com/). It
is in yaml format. The top level structure is an array of [jobs](https://github.com/hyperledger/burrow/blob/main/deploy/def/job.go).
The different job types are [defined here](https://github.com/hyperledger/burrow/blob/main/deploy/def/jobs.go).

You can invoke burrow from the command line:

```shell
burrow deploy -a CF8F9480252B70D59CF5B5F3CAAA75FEAF6A4B33 deploy.yaml
```

Each job in the playbook has a name. This name can be used in later jobs to refer to the result of a previous job (e.g. the address of a contract
which was deployed). The jobs are executed in-order.

Whenever an account needs to be specified, the key name in the burrow keys server can also be used.

## Deploy

The deploy job compiles a solidity source file to a bin file which is then deployed to the chain. This type of job has the following
parameters:

* _source:_ the input address from which to do the deploy transaction
* _contract:_ the path to the solidity source file
* _instance:_ once solidity source file can contain multiple contracts. This field is ignored if there is only one contract in the
  source. If there are multiple, the contract must match the filename, else this field. If this field is set to "all", all contracts
  in will be deployed.
* _libraries:_ list of the library address to link against
* _data:_ the arguments to the contract's constructor

The solidity source file is compiled using the [solidity compiler](https://github.com/ethereum/solidity) unless the `--wasm` argument was given
on the burrow deploy command line, in which case the [solang compiler](https://github.com/hyperledger-labs/solang) is used.

The contract is deployed with its metadata, so that we can retrieve the ABI when we need to call a function of this contract. For this
reason, the bin file is a modified version of the [solidity output json](https://solidity.readthedocs.io/en/v0.5.11/using-the-compiler.html#output-description).

A solidity source file can have any number of contracts, and those contract names do not have to match the file name of the source. The resulting bin
file(s) is named according to the name of the contract(s). To select which contracts to use, specifiy the _instance_ field.

If the _contract_ is specified as a bin file, compilation will be skipped. It can be useful to separate compilation from deployment using the build job,
which is described next.

## Build

The build job is used to only compile solidity and do not do any deployment. This only has one parameter:

* _contract:_ the path to the solidity source

## Call / Query-Contract

The call and query contract job is for executing contract code by way of running one of the functions. The call job will create a transaction
and will have to wait until the next block to retrieve the result; the query-contract job is for accessing read-only functions which do not require
write access. This type of job has the following parameters:

* _source:_ the input address which should execute the transaction
* _destination:_ the account to access
* _function:_ the name of the function to call
* _data:_ the arguments to the function call
* _bin:_ the path to the abi or bin file

The _destination_ field can either be a hex encoding adress, the name of a key name (e.g. Root\_0 or Participant\_1), or the result
of previous burrow deploy job. In the latter case the name of the job must be specified, prefixed with $.

If the contract was deployed without metadata (e.g. using the burrow js module or with an earlier version of burrow deploy) the abi must be
specified. This must be the path to the contract bin file or abi file.

## Proposal

This is described in the [proposal tutorial](tutorials/8-proposals.md).
