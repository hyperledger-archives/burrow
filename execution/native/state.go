package native

import (
	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
)

// This wrapper provides a state that behaves 'as if' the natives were stored directly in state.
// TODO: we may want to actually store native sentinels (and thier metadata) in on-disk state down the line
type State struct {
	natives *Natives
	backend acmstate.ReaderWriter
}

// Get a new state that wraps the backend but intercepts any calls to natives returning appropriate errors message
// or an Account sentinel for the particular native
func NewState(natives *Natives, backend acmstate.ReaderWriter) *State {
	return &State{
		natives: natives,
		backend: backend,
	}
}

func (s *State) UpdateAccount(updatedAccount *acm.Account) error {
	err := s.ensureNonNative(updatedAccount.Address, "update account")
	if err != nil {
		return err
	}
	return s.backend.UpdateAccount(updatedAccount)
}

func (s *State) GetAccount(address crypto.Address) (*acm.Account, error) {
	contract := s.natives.GetByAddress(address)
	if contract != nil {
		return account(contract), nil
	}
	return s.backend.GetAccount(address)
}

func (s *State) RemoveAccount(address crypto.Address) error {
	err := s.ensureNonNative(address, "remove account")
	if err != nil {
		return err
	}
	return s.backend.RemoveAccount(address)
}

func (s *State) GetStorage(address crypto.Address, key binary.Word256) ([]byte, error) {
	err := s.ensureNonNative(address, "get storage")
	if err != nil {
		return nil, err
	}
	return s.backend.GetStorage(address, key)
}

func (s *State) SetStorage(address crypto.Address, key binary.Word256, value []byte) error {
	err := s.ensureNonNative(address, "set storage")
	if err != nil {
		return err
	}
	return s.backend.SetStorage(address, key, value)
}

func (s *State) GetMetadata(metahash acmstate.MetadataHash) (string, error) {
	return s.backend.GetMetadata(metahash)
}

func (s *State) SetMetadata(metahash acmstate.MetadataHash, metadata string) error {
	return s.backend.SetMetadata(metahash, metadata)
}

func (s *State) ensureNonNative(address crypto.Address, action string) error {
	contract := s.natives.GetByAddress(address)
	if contract != nil {
		return errors.Errorf(errors.Code.ReservedAddress,
			"cannot %s at %v because that address is reserved for a native contract '%s'",
			action, address, contract.FullName())
	}
	return nil
}

func account(callable Native) *acm.Account {
	return &acm.Account{
		Address:    callable.Address(),
		NativeName: callable.FullName(),
		// TODO: this is not populated yet, see FIXME note on native.Contract
		ContractMeta: callable.ContractMeta(),
	}
}
