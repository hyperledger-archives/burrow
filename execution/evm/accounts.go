package evm

import (
	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/logging"
	ptypes "github.com/hyperledger/burrow/permission/types"
)

// Create a new account from a parent 'creator' account. The creator account will have its
// sequence number incremented
func DeriveNewAccount(creator *acm.Account, permissions ptypes.AccountPermissions,
	logger *logging.Logger) *acm.Account {
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
	acc := acm.NewContractAccount(addr, permissions)
	return acc
}
