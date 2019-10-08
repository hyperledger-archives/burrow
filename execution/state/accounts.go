package state

import (
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/encoding"
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
	account := new(acm.Account)
	err = encoding.Decode(accBytes, account)
	if err != nil {
		return nil, fmt.Errorf("could not decode Account: %v", err)
	}
	return account, nil
}

func (ws *writeState) statsAddAccount(acc *acm.Account) {
	if acc != nil {
		if len(acc.EVMCode) > 0 || len(acc.WASMCode) > 0 {
			ws.accountStats.AccountsWithCode++
		} else {
			ws.accountStats.AccountsWithoutCode++
		}
	}
}

func (ws *writeState) statsRemoveAccount(acc *acm.Account) {
	if acc != nil {
		if len(acc.EVMCode) > 0 || len(acc.WASMCode) > 0 {
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
	bs, err := encoding.Encode(account)
	if err != nil {
		return fmt.Errorf("UpdateAccount could not encode account: %v", err)
	}
	tree, err := ws.forest.Writer(keys.Account.Prefix())
	if err != nil {
		return err
	}
	updated := tree.Set(keys.Account.KeyNoPrefix(account.Address), bs)
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
		account := new(acm.Account)
		err := encoding.Decode(accBytes, account)
		if err != nil {
			return err
		}
		ws.statsRemoveAccount(account)
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
		account := new(acm.Account)
		err := encoding.Decode(value, account)
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

func (s *ReadState) GetStorage(address crypto.Address, key binary.Word256) ([]byte, error) {
	keyFormat := keys.Storage.Fix(address)
	tree, err := s.Forest.Reader(keyFormat.Prefix())
	if err != nil {
		return []byte{}, err
	}
	return tree.Get(keyFormat.KeyNoPrefix(key)), nil
}

func (ws *writeState) SetStorage(address crypto.Address, key binary.Word256, value []byte) error {
	keyFormat := keys.Storage.Fix(address)
	tree, err := ws.forest.Writer(keyFormat.Prefix())
	if err != nil {
		return err
	}
	zero := true
	for _, b := range value {
		if b != 0 {
			zero = false
			break
		}
	}
	if zero {
		tree.Delete(keyFormat.KeyNoPrefix(key))
	} else {
		tree.Set(keyFormat.KeyNoPrefix(key), value)
	}
	return nil
}

func (s *ReadState) IterateStorage(address crypto.Address, consumer func(key binary.Word256, value []byte) error) error {
	keyFormat := keys.Storage.Fix(address)
	tree, err := s.Forest.Reader(keyFormat.Prefix())
	if err != nil {
		return err
	}
	return tree.Iterate(nil, nil, true,
		func(key []byte, value []byte) error {

			if len(key) != binary.Word256Bytes {
				return fmt.Errorf("key '%X' stored for account %s is not a %v-byte word",
					key, address, binary.Word256Bytes)
			}
			if len(value) != binary.Word256Bytes {
				return fmt.Errorf("value '%X' stored for account %s is not a %v-byte word",
					key, address, binary.Word256Bytes)
			}
			return consumer(binary.LeftPadWord256(key), value)
		})
}
