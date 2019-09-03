# Transactions

Burrow implements a number of transaction types. Transactions will be ordered by our consensus mechanism (Tendermint) and applied to our application state machine - replicated across all Burrow nodes. Each transaction is applied atomically and runs deterministically. The transactions contain the arguments for an [execution context](/execution/contexts).

Our transactions are defined in Protobuf [here](/protobuf/payload.proto).

Transactions can be built using our GRPC client libraries programmatically, via [burrow.js](/docs/js-api.md), or with`burrow deploy`, see our [deploying contracts guide](/docs/tutorials/3-deploy-contracts.md).

## [CallTx](https://godoc.org/github.com/hyperledger/burrow/txs/payload#CallTx)

Our core transaction type that calls EVM code, possibly transferring value. It takes the following parameters:

| Parameter | Type | Description |
| ----------|------|-------------|
| Input | 
	Address *github_com_hyperledger_burrow_crypto.Address 
	GasLimit uint64 
	// Fee to offer validators for processing transaction
	Fee uint64 `protobuf:"varint,4,opt,name=Fee,proto3" json:"Fee,omitempty"`
	// EVM bytecode
	Data github_com_hyperledger_burrow_binary.HexBytes `protobuf:"bytes,5,opt,name=Data,proto3,customtype=github.com/hyperledger/burrow/binary.HexBytes" json:"Data"`
	// WASM bytecode
	WASM github_com_hyperledger_burrow_binary.HexBytes `protobuf:"bytes,6,opt,name=WASM,proto3,customtype=github.com/hyperledger/burrow/binary.HexBytes" json:"tags,omitempty"`
	// Set of contracts this code will deploy
	ContractMeta         []*ContractMeta `protobuf:"bytes,7,rep,name=ContractMeta,proto3" json:"ContractMeta,omitempty"`

## [SendTx](https://godoc.org/github.com/hyperledger/burrow/txs/payload#SendTx)

Allows [native token](/docs/reference/accounts.md) to be sent from multiple inputs to multiple outputs. The basic value transfer function that calls no EVM Code.

## [NameTx](https://godoc.org/github.com/hyperledger/burrow/txs/payload#NameTx)

Provides access to a global name registry service that associates a particular string key with a data payload and an owner. The control of the name is guaranteed for the period of the lease which is a determined by a fee.

Note: a future revision will change the way in which leases are calculated. Currently we use a somewhat historically-rooted fixed fee, see the [`NameCostPerBlock` function](/execution/names/names.go).

## [BondTx](https://godoc.org/github.com/hyperledger/burrow/txs/payload#BondTx)

This allows validators nominate themselves to the validator set by placing a bond subtracted from their balance.

## [UnbondTx](https://godoc.org/github.com/hyperledger/burrow/txs/payload#UnbondTx)

This allows validators remove themselves to the validator set returning their bond to their balance.

## [BatchTx](https://godoc.org/github.com/hyperledger/burrow/txs/payload#BatchTx)

Runs a set of transactions atomically in a single meta-transaction within a single block

## [GovTx](https://godoc.org/github.com/hyperledger/burrow/txs/payload#GovTx)

An all-powerful transaction for modifying existing accounts.

## [ProposalTx](https://godoc.org/github.com/hyperledger/burrow/txs/payload#ProposalTx)

A transaction type containing a batch of transactions on which a ballot is held to determine whether to execute, see [proposals tutorial](/docs/tutorials/8-proposals.md)

## [PermsTx](https://godoc.org/github.com/hyperledger/burrow/txs/payload#PermsTx)

A transaction to modify the permissions of accounts.

