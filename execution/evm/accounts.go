package evm

import (
	acm "github.com/hyperledger/burrow/account"
	ptypes "github.com/hyperledger/burrow/permission/types"
)

// Create a new account from a parent 'creator' account. The creator account will have its
// sequence number incremented
func DeriveNewAccount(creator acm.MutableAccount, permissions ptypes.AccountPermissions) acm.MutableAccount {
	// Generate an address
	sequence := creator.Sequence()
	creator.IncSequence()

	addr := acm.NewContractAddress(creator.Address(), sequence)

	// Create account from address.
	return (&acm.ConcreteAccount{
		Address:     addr,
		Balance:     0,
		Code:        nil,
		Sequence:    0,
		Permissions: permissions,
	}).MutableAccount()
}
