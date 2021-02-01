package engine

import (
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/permission"
)

type Maybe interface {
	PushError(err error) bool
	Error() error
}

func GetAccount(st acmstate.Reader, m Maybe, address crypto.Address) *acm.Account {
	acc, err := st.GetAccount(address)
	if err != nil {
		m.PushError(err)
		return nil
	}
	return acc
}

// Guaranteed to return a non-nil account, if the account does not exist returns a pointer to the zero-value of Account
// and pushes an error.
func MustGetAccount(st acmstate.Reader, m Maybe, address crypto.Address) *acm.Account {
	acc := GetAccount(st, m, address)
	if acc == nil {
		m.PushError(errors.Errorf(errors.Codes.NonExistentAccount, "account %v does not exist", address))
		return &acm.Account{}
	}
	return acc
}

func EnsurePermission(callFrame *CallFrame, address crypto.Address, perm permission.PermFlag) error {
	hasPermission, err := HasPermission(callFrame, address, perm)
	if err != nil {
		return err
	} else if !hasPermission {
		return errors.PermissionDenied{
			Address: address,
			Perm:    perm,
		}
	}
	return nil
}

// CONTRACT: it is the duty of the contract writer to call known permissions
// we do not convey if a permission is not set
// (unlike in state/execution, where we guarantee HasPermission is called
// on known permissions and panics else)
// If the perm is not defined in the acc nor set by default in GlobalPermissions,
// this function returns false.
func HasPermission(st acmstate.Reader, address crypto.Address, perm permission.PermFlag) (bool, error) {
	acc, err := st.GetAccount(address)
	if err != nil {
		return false, err
	}
	if acc == nil {
		return false, fmt.Errorf("account %v does not exist", address)
	}
	globalPerms, err := acmstate.GlobalAccountPermissions(st)
	if err != nil {
		return false, err
	}
	perms := acc.Permissions.Base.Compose(globalPerms.Base)
	value, err := perms.Get(perm)
	if err != nil {
		return false, err
	}
	return value, nil
}

func CreateAccount(st acmstate.ReaderWriter, address crypto.Address) error {
	acc, err := st.GetAccount(address)
	if err != nil {
		return err
	}
	if acc != nil {
		if acc.NativeName != "" {
			return errors.Errorf(errors.Codes.ReservedAddress,
				"cannot create account at %v because that address is reserved for a native contract '%s'",
				address, acc.NativeName)
		}
		return errors.Errorf(errors.Codes.DuplicateAddress,
			"tried to create an account at an address that already exists: %v", address)
	}
	return st.UpdateAccount(&acm.Account{Address: address})
}

func AddressFromName(name string) (address crypto.Address) {
	hash := crypto.Keccak256([]byte(name))
	copy(address[:], hash[len(hash)-crypto.AddressLength:])
	return
}
