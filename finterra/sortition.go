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
	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/txs"
)

type ITransactor interface {
	BroadcastTx(tx txs.Tx) (*txs.Receipt, error)
}

type Sortition struct {
	acm.Addressable
	acm.Signer
	vrf           VRF
	transactor    ITransactor
	chainID       string
	validatorPool ValidatorPoolIterable
	logger        *logging.Logger
}

func NewSortition(addressable acm.Addressable, signer acm.Signer, chainID string, logger *logging.Logger) acm.Sortition {
	return &Sortition{
		Addressable: addressable,
		Signer:      signer,
		vrf:         NewVRF(addressable, signer),
		chainID:     chainID,
		logger:      logger,
		transactor:  nil,
	}
}

func (s *Sortition) SetTransactor(transactor ITransactor) {
	s.transactor = transactor
}

func (s *Sortition) SetValidaorPool(validatorPool ValidatorPoolIterable) {
	s.validatorPool = validatorPool
}

// Evaluate return the vrf for self chossing to be a validator
func (s *Sortition) Evaluate(blockHeight uint64, prevBlockHash []byte) {
	totalStake, valStake := s.GetTotalStake(s.Address())

	s.vrf.SetMax(totalStake)

	index, proof := s.vrf.Evaluate(prevBlockHash)

	if index < valStake {

		s.logger.InfoMsg("This validator is choosen to be in set at height %v", blockHeight)

		tx := txs.NewSortitionTx(
			s.PublicKey(),
			blockHeight,
			index,
			proof)

		tx.Signature, _ = acm.ChainSign(s, s.chainID, tx)

		if s.transactor != nil {
			s.transactor.BroadcastTx(tx)
		}
	}
}

func (s *Sortition) Verify(prevBlockHash []byte, publicKey acm.PublicKey, index uint64, proof []byte) bool {

	totalStake, valStake := s.GetTotalStake(publicKey.Address())

	// Note: totalStake can be changed by time on verifying
	// So we calculate the index again
	s.vrf.SetMax(totalStake)

	index2, result := s.vrf.Verify(prevBlockHash, publicKey, proof)

	if result == false {
		return false
	}

	return index2 < valStake
}

func (s *Sortition) GetAddress() acm.Address {
	return s.Address()
}

func (s *Sortition) GetTotalStake(address acm.Address) (totalStake uint64, validatorStake uint64) {
	totalStake = 0
	validatorStake = 0

	s.validatorPool.IterateValidator(func(validator acm.Validator) (stop bool) {
		totalStake += validator.Stake()

		if address == validator.Address() {
			validatorStake = validator.Stake()
		}

		return false
	})

	return
}
