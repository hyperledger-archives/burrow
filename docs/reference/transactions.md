# Transactions

Burrow implements a number of transaction types. Transactions will be ordered by our consensus mechanism (Tendermint) and applied to our application state machine - 
replicated across all Burrow nodes. Each transaction is applied atomically and runs deterministically. The transactions contain the arguments for an 
[execution context](https://github.com/hyperledger/burrow/tree/master/execution/contexts).

Our transactions are defined in Protobuf [here](https://github.com/hyperledger/burrow/blob/master/protobuf/payload.proto).

Transactions can be built using our GRPC client libraries programmatically, via [burrow.js](js-api.md), or with `burrow deploy` - see our [deployment guide](deploy.md).

## TxInput

| Parameter | Type | Description |
| ----------|------|-------------|
| Address | Address | The address of an account issuing this transaction - the transaction envelope must also be signed by the private key associated with this address |
| Amount | uint64 | The amount of native token to transfer from the input to the output of the transaction |
| Sequence | uint64 | A counter that must match the current value of the input account's Sequence plus one - i.e. the Sequence must equal n if this is the nth transaction issued by this account |


## CallTx

Our core transaction type that calls EVM code, possibly transferring value. It takes the following parameters:

| Parameter | Type | Description |
| ----------|------|-------------|
| Input | TxInput | The external 'caller' account - will be the initial SENDER and CALLER |
| Address | *Address | The address 'callee' contract - the contract whose code will be executed. If this value is nil then the CallTx is interpreted as contract creation and will deploy the bytecode contained in Data or WASM |
| GasLimit | uint64 | The maximum number of computational steps that we will allow to run before aborted the transaction execution. Measured according to our hardcoded simplified gas schedule (one gas unit per operation). Ensure transaction termination. If 0 a default cap will be used. |
| Fee | uint64 | An optional fee to be subtracted from the input amount - currently this fee is simply burnt! In the future fees will be collected and disbursed amongst validators as part of our token economics system |
| Data | []byte |  If the CallTx is a deployment (i.e. Address is nil) then this data will be executed as EVM bytecode will and the return value will be used to instatiate a new contract. If the CallTx is a plain call then the data will form the input tape for the EVM call |

## SendTx

Allows [native token](reference/participants.md) to be sent from multiple inputs to multiple outputs. The basic value transfer function that calls no EVM Code.

## NameTx

Provides access to a global name registry service that associates a particular string key with a data payload and an owner. The control of the name is guaranteed for 
the period of the lease which is a determined by a fee.

> A future revision will change the way in which leases are calculated. Currently we use a somewhat historically-rooted fixed fee, see the [`NameCostPerBlock` function](https://github.com/hyperledger/burrow/blob/master/execution/names/names.go#L83).

## BondTx

This allows validators nominate themselves to the validator set by placing a bond subtracted from their balance.

For more information see the [bonding documentation](reference/bonding.md).

## UnbondTx

This allows validators remove themselves to the validator set returning their bond to their balance.

## BatchTx

Runs a set of transactions atomically in a single meta-transaction within a single block

## GovTx

An all-powerful transaction for modifying existing accounts.

## ProposalTx

A transaction type containing a batch of transactions on which a ballot is held to determine whether to execute, see the [proposals tutorial](tutorials/8-proposals.md).

## PermsTx

A transaction to modify the permissions of accounts.

