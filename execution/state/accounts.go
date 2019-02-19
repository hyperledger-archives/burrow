package state

import (
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
)

// Returns nil if account does not exist with given address.
func (s *ReadState) GetAccount(address crypto.Address) (*acm.Account, error) {
	tree, err := s.Forest.Reader(keys.Account.Prefix())
	if err != nil {
		return nil, err
	}
	accBytes := tree.Get(keys.Account.KeyNoPrefix(address))
	if accBytes == nil {
		return nil, nil
	}
	return acm.Decode(accBytes)
}

func (ws *writeState) statsAddAccount(acc *acm.Account) {
	if acc != nil {
		if len(acc.Code) > 0 {
			ws.accountStats.AccountsWithCode++
		} else {
			ws.accountStats.AccountsWithoutCode++
		}
	}
}

func (ws *writeState) statsRemoveAccount(acc *acm.Account) {
	if acc != nil {
		if len(acc.Code) > 0 {
			ws.accountStats.AccountsWithCode--
		} else {
			ws.accountStats.AccountsWithoutCode--
		}
	}
}

func (ws *writeState) UpdateAccount(account *acm.Account) error {
	if account == nil {
		return fmt.Errorf("UpdateAccount passed nil account in State")
	}
	encodedAccount, err := account.Encode()
	if err != nil {
		return fmt.Errorf("UpdateAccount could not encode account: %v", err)
	}
	tree, err := ws.forest.Writer(keys.Account.Prefix())
	if err != nil {
		return err
	}
	updated := tree.Set(keys.Account.KeyNoPrefix(account.Address), encodedAccount)
	if updated {
		ws.statsAddAccount(account)
	}
	return nil
}

func (ws *writeState) RemoveAccount(address crypto.Address) error {
	tree, err := ws.forest.Writer(keys.Account.Prefix())
	if err != nil {
		return err
	}
	accBytes, deleted := tree.Delete(keys.Account.KeyNoPrefix(address))
	if deleted {
		acc, err := acm.Decode(accBytes)
		if err != nil {
			return err
		}
		ws.statsRemoveAccount(acc)
		// Delete storage associated with account too
		_, err = ws.forest.Delete(keys.Storage.Key(address))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ReadState) IterateAccounts(consumer func(*acm.Account) error) error {
	tree, err := s.Forest.Reader(keys.Account.Prefix())
	if err != nil {
		return err
	}
	return tree.Iterate(nil, nil, true, func(key []byte, value []byte) error {
		account, err := acm.Decode(value)
		if err != nil {
			return fmt.Errorf("IterateAccounts could not decode account: %v", err)
		}
		return consumer(account)
	})
}

func (s *State) GetAccountStats() acmstate.AccountStats {
	return s.writeState.accountStats
}

// Storage

func (s *ReadState) GetStorage(address crypto.Address, key binary.Word256) (binary.Word256, error) {
	keyFormat := keys.Storage.Fix(address)
	tree, err := s.Forest.Reader(keyFormat.Prefix())
	if err != nil {
		return binary.Zero256, err
	}
	return binary.LeftPadWord256(tree.Get(keyFormat.KeyNoPrefix(key))), nil
}

func (ws *writeState) SetStorage(address crypto.Address, key, value binary.Word256) error {
	keyFormat := keys.Storage.Fix(address)
	tree, err := ws.forest.Writer(keyFormat.Prefix())
	if err != nil {
		return err
	}
	if value == binary.Zero256 {
		tree.Delete(keyFormat.KeyNoPrefix(key))
	} else {
		tree.Set(keyFormat.KeyNoPrefix(key), value.Bytes())
	}
	return nil
}

func (s *ReadState) IterateStorage(address crypto.Address, consumer func(key, value binary.Word256) error) error {
	keyFormat := keys.Storage
	tree, err := s.Forest.Reader(keyFormat.Prefix())
	if err != nil {
		return err
	}
	return tree.Iterate(nil, nil, true,
		func(key []byte, value []byte) error {

			if len(key) != binary.Word256Length {
				return fmt.Errorf("key '%X' stored for account %s is not a %v-byte word",
					key, address, binary.Word256Length)
			}
			if len(value) != binary.Word256Length {
				return fmt.Errorf("value '%X' stored for account %s is not a %v-byte word",
					key, address, binary.Word256Length)
			}
			return consumer(binary.LeftPadWord256(key), binary.LeftPadWord256(value))
		})
}
