// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package account

import (
	"fmt"

	"github.com/tendermint/go-wire"
)

type AddressableSigner interface {
	Addressable
	Signer
}

type PrivateAccount interface {
	AddressableSigner
	PrivateKey() PrivateKey
}

//
type ConcretePrivateAccount struct {
	Address    Address
	PublicKey  PublicKey
	PrivateKey PrivateKey
}

type concretePrivateAccountWrapper struct {
	*ConcretePrivateAccount `json:"unwrap"`
}

var _ PrivateAccount = concretePrivateAccountWrapper{}

var _ = wire.RegisterInterface(struct{ PrivateAccount }{}, wire.ConcreteType{concretePrivateAccountWrapper{}, 0x01})

func AsConcretePrivateAccount(privateAccount PrivateAccount) *ConcretePrivateAccount {
	if privateAccount == nil {
		return nil
	}
	// Avoid a copy
	if ca, ok := privateAccount.(concretePrivateAccountWrapper); ok {
		return ca.ConcretePrivateAccount
	}
	return &ConcretePrivateAccount{
		Address:    privateAccount.Address(),
		PublicKey:  privateAccount.PublicKey(),
		PrivateKey: privateAccount.PrivateKey(),
	}
}

func (cpaw concretePrivateAccountWrapper) Address() Address {
	return cpaw.ConcretePrivateAccount.Address
}

func (cpaw concretePrivateAccountWrapper) PublicKey() PublicKey {
	return cpaw.ConcretePrivateAccount.PublicKey
}

func (cpaw concretePrivateAccountWrapper) PrivateKey() PrivateKey {
	return cpaw.ConcretePrivateAccount.PrivateKey
}

func (cpaw concretePrivateAccountWrapper) String() string {
	return cpaw.ConcretePrivateAccount.String()
}

// ConcretePrivateAccount

func (pa ConcretePrivateAccount) PrivateAccount() PrivateAccount {
	return concretePrivateAccountWrapper{ConcretePrivateAccount: &pa}
}

func (pa ConcretePrivateAccount) Sign(msg []byte) (Signature, error) {
	return pa.PrivateKey.Sign(msg)
}

func ChainSign(signer Signer, chainID string, o Signable) (Signature, error) {
	sig, err := signer.Sign(SignBytes(chainID, o))
	if err != nil {
		return Signature{}, err
	}
	return sig, nil
}

func (pa *ConcretePrivateAccount) String() string {
	return fmt.Sprintf("ConcretePrivateAccount{%s}", pa.Address)
}

// Convert slice of ConcretePrivateAccounts to slice of SigningAccounts
func SigningAccounts(concretePrivateAccounts []*ConcretePrivateAccount) []AddressableSigner {
	signingAccounts := make([]AddressableSigner, len(concretePrivateAccounts))
	for i, cpa := range concretePrivateAccounts {
		signingAccounts[i] = cpa.PrivateAccount()
	}
	return signingAccounts
}

// Generates a new account with private key.
func GeneratePrivateAccount() (PrivateAccount, error) {
	privateKey, err := GeneratePrivateKey(nil)
	if err != nil {
		return nil, err
	}
	publicKey := privateKey.PublicKey()
	return ConcretePrivateAccount{
		Address:    publicKey.Address(),
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}.PrivateAccount(), nil
}

// Generates a new account with private key from SHA256 hash of a secret
func GeneratePrivateAccountFromSecret(secret string) PrivateAccount {
	privateKey := PrivateKeyFromSecret(secret)
	publicKey := privateKey.PublicKey()
	return ConcretePrivateAccount{
		Address:    publicKey.Address(),
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}.PrivateAccount()
}

func GeneratePrivateAccountFromPrivateKeyBytes(privKeyBytes []byte) (PrivateAccount, error) {
	privateKey, err := Ed25519PrivateKeyFromRawBytes(privKeyBytes)
	if err != nil {
		return nil, err
	}
	publicKey := privateKey.PublicKey()
	return ConcretePrivateAccount{
		Address:    publicKey.Address(),
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}.PrivateAccount(), nil
}
