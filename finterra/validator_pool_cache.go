// MIT License
//
// Copyright (c) 2018 Finterra
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//

package finterra

import (
	"errors"
	"sync"

	acm "github.com/hyperledger/burrow/account"
	bcm "github.com/hyperledger/burrow/blockchain"
)

var (
	ErrValidatorChanged = errors.New("Validator has changed before in this height")
)

type ValidatorStatus int

const (
	AddToPool      ValidatorStatus = iota /// Bonding transaction
	RemoveFromPool                        /// Unbonding transaction
	AddToSet                              /// Sortition transaction
	RemoveFromSet                         /// No transaction for this state change, but it is deterministic and enforce by the protocol
	Update                                /// No transaction for this state change, but it is deterministic and enforce by the protocol
)

type ValidatorPoolCache struct {
	sync.RWMutex
	backend ValidatorPoolGetter
	changes map[acm.Validator]ValidatorStatus
}

func NewValidatorPoolCache(backend ValidatorPoolGetter) *ValidatorPoolCache {
	return &ValidatorPoolCache{
		backend: backend,
		changes: make(map[acm.Validator]ValidatorStatus),
	}
}

func (vpc *ValidatorPoolCache) GetValidator(address acm.Address) acm.Validator {
	vpc.Lock()
	defer vpc.Unlock()

	for v := range vpc.changes {
		if v.Address() == address {
			return v
		}
	}

	return vpc.backend.GetValidator(address)
}

func (vpc *ValidatorPoolCache) GetMutableValidator(address acm.Address) acm.MutableValidator {
	validator := vpc.GetValidator(address)
	if validator == nil {
		return nil
	}
	return acm.AsMutableValidator(validator)
}

func (vpc *ValidatorPoolCache) Reset() {
	vpc.Lock()
	defer vpc.Unlock()
	for k := range vpc.changes {
		delete(vpc.changes, k)
	}
}

// Syncs the NameRegCache and Resets it to use NameRegWriter as the backend NameRegGetter
func (vpc *ValidatorPoolCache) Flush(state ValidatorPoolWriter, validatorSet bcm.ValidatorSet) error {
	vpc.Lock()
	defer vpc.Unlock()
	for val, status := range vpc.changes {
		switch status {
		case AddToSet:
			if err := validatorSet.JoinToTheSet(val); err != nil {
				return err
			}

		case RemoveFromSet:
			if err := validatorSet.LeaveFromTheSet(val); err != nil {
				return err
			}

		case Update, AddToPool:
			if err := state.UpdateValidator(val); err != nil {
				return err
			}

		case RemoveFromPool:
			if err := validatorSet.LeaveFromTheSet(val); err != nil {
				return err
			}

			if err := state.RemoveValidator(val); err != nil {
				return err
			}
		}
	}

	return nil
}

func (vpc *ValidatorPoolCache) AddToPool(validator acm.Validator) error {
	vpc.Lock()
	defer vpc.Unlock()
	if vpc.hasChanged(validator) {
		return ErrValidatorChanged
	}

	vpc.changes[validator] = AddToPool
	return nil
}

func (vpc *ValidatorPoolCache) AddToSet(validator acm.Validator) error {
	vpc.Lock()
	defer vpc.Unlock()
	if vpc.hasChanged(validator) {
		return ErrValidatorChanged
	}

	vpc.changes[validator] = AddToSet
	return nil
}

func (vpc *ValidatorPoolCache) RemoveFromPool(validator acm.Validator) error {
	vpc.Lock()
	defer vpc.Unlock()
	if vpc.hasChanged(validator) {
		return ErrValidatorChanged
	}

	vpc.changes[validator] = RemoveFromPool
	return nil
}

func (vpc *ValidatorPoolCache) RemoveFromSet(validator acm.Validator) error {
	vpc.Lock()
	defer vpc.Unlock()
	if vpc.hasChanged(validator) {
		return ErrValidatorChanged
	}

	vpc.changes[validator] = RemoveFromSet
	return nil
}

func (vpc *ValidatorPoolCache) UpdateValidator(validator acm.Validator) error {
	vpc.Lock()
	defer vpc.Unlock()
	if vpc.hasChanged(validator) {
		return ErrValidatorChanged
	}

	vpc.changes[validator] = Update
	return nil
}

/// In each height, Validator state can be changed once.
/// We check here if it has chenged once or not
func (vpc *ValidatorPoolCache) hasChanged(validator acm.Validator) bool {
	if _, ok := vpc.changes[validator]; ok {
		return true
	}

	return false
}

func (vpc *ValidatorPoolCache) ValidatorCount() int {
	return vpc.backend.ValidatorCount()
}
