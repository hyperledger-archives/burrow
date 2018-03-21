package evm

import (
	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	ptypes "github.com/hyperledger/burrow/permission/types"
)

// Create a new account from a parent 'creator' account. The creator account will have its
// sequence number incremented
func DeriveNewAccount(creator acm.MutableAccount, permissions ptypes.AccountPermissions,
	logger logging_types.InfoTraceLogger) acm.MutableAccount {
	// Generate an address
	sequence := creator.Sequence()
	logging.TraceMsg(logger, "Incrementing sequence number in DeriveNewAccount()",
		"tag", "sequence",
		"account", creator.Address(),
		"old_sequence", sequence,
		"new_sequence", sequence+1)
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
