package evm

import (
	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/permission"
)

// Create a new account from a parent 'creator' account. The creator account will have its
// sequence number incremented
func DeriveNewAccount(creator *acm.MutableAccount, permissions permission.AccountPermissions,
	logger *logging.Logger) *acm.MutableAccount {
	// Generate an address
	sequence := creator.Sequence()
	logger.TraceMsg("Incrementing sequence number in DeriveNewAccount()",
		"tag", "sequence",
		"account", creator.Address(),
		"old_sequence", sequence,
		"new_sequence", sequence+1)
	creator.IncSequence()

	addr := crypto.NewContractAddress(creator.Address(), sequence)

	// Create account from address.
	return acm.ConcreteAccount{
		Address:     addr,
		Balance:     0,
		Code:        nil,
		Sequence:    0,
		Permissions: permissions,
	}.MutableAccount()
}
