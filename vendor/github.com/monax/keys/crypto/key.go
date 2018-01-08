/*
	This file is part of go-ethereum
	go-ethereum is free software: you can redistribute it and/or modify
	it under the terms of the GNU Lesser General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-ethereum is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU Lesser General Public License
	along with go-ethereum.  If not, see <http://www.gnu.org/licenses/>.
*/
/**
 * @authors
 * 	Gustav Simonsson <gustav.simonsson@gmail.com>
 *	Ethan Buchman <ethan@erisindustries.com> (adapt for ed25519 keys also)
 * @date 2015
 *
 */

package crypto

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/monax/keys/crypto/randentropy"
	"github.com/monax/keys/crypto/secp256k1"
	"github.com/tendermint/ed25519"

	"github.com/monax/keys/crypto/helpers"
	uuid "github.com/wayn3h0/go-uuid"
)

type InvalidCurveErr string

func (err InvalidCurveErr) Error() string {
	return fmt.Sprintf("invalid curve type %v", err)
}

type NoPrivateKeyErr string

func (err NoPrivateKeyErr) Error() string {
	return fmt.Sprintf("Private key is not available or is encrypted")
}

type KeyType struct {
	CurveType CurveType
	AddrType  AddrType
}

func (typ KeyType) String() string {
	return fmt.Sprintf("%s,%s", typ.CurveType.String(), typ.AddrType.String())
}

func KeyTypeFromString(s string) (k KeyType, err error) {
	spl := strings.Split(s, ",")
	if len(spl) != 2 {
		return k, fmt.Errorf("KeyType should be (CurveType,AddrType)")
	}

	cType, aType := spl[0], spl[1]
	if k.CurveType, err = CurveTypeFromString(cType); err != nil {
		return
	}
	k.AddrType, err = AddrTypeFromString(aType)
	return
}

//-----------------------------------------------------------------------------
// curve type

type CurveType uint8

func (k CurveType) String() string {
	switch k {
	case CurveTypeSecp256k1:
		return "secp256k1"
	case CurveTypeEd25519:
		return "ed25519"
	default:
		return "unknown"
	}
}

func CurveTypeFromString(s string) (CurveType, error) {
	switch s {
	case "secp256k1":
		return CurveTypeSecp256k1, nil
	case "ed25519":
		return CurveTypeEd25519, nil
	default:
		var k CurveType
		return k, InvalidCurveErr(s)
	}
}

const (
	CurveTypeSecp256k1 CurveType = iota
	CurveTypeEd25519
)

//-----------------------------------------------------------------------------
// address type

type AddrType uint8

func (a AddrType) String() string {
	switch a {
	case AddrTypeRipemd160:
		return "ripemd160"
	case AddrTypeRipemd160Sha256:
		return "ripemd160sha256"
	case AddrTypeSha3:
		return "sha3"
	default:
		return "unknown"
	}
}

func AddrTypeFromString(s string) (AddrType, error) {
	switch s {
	case "ripemd160":
		return AddrTypeRipemd160, nil
	case "ripemd160sha256":
		return AddrTypeRipemd160Sha256, nil
	case "sha3":
		return AddrTypeSha3, nil
	default:
		var a AddrType
		return a, fmt.Errorf("unknown addr type %s", s)
	}
}

const (
	AddrTypeRipemd160 AddrType = iota
	AddrTypeRipemd160Sha256
	AddrTypeSha3
)

func AddressFromPub(addrType AddrType, pub []byte) (addr []byte) {
	switch addrType {
	case AddrTypeRipemd160:
		// let tendermint/binary handle because
		// it encodes the type byte ...
	case AddrTypeRipemd160Sha256:
		addr = Ripemd160(Sha256(pub))
	case AddrTypeSha3:
		addr = Sha3(pub[1:])[12:]
	}
	return
}

//-----------------------------------------------------------------------------
// main key struct and functions (sign, pubkey, verify)

type Key struct {
	Id         uuid.UUID // Version 4 "random" for unique id not derived from key data
	Type       KeyType   // contains curve and addr types
	Address    []byte    // reference id
	PrivateKey []byte    // we don't store pub
}

func NewKey(typ KeyType) (*Key, error) {
	switch typ.CurveType {
	case CurveTypeSecp256k1:
		return newKeySecp256k1(typ.AddrType), nil
	case CurveTypeEd25519:
		return newKeyEd25519(typ.AddrType), nil
	default:
		return nil, fmt.Errorf("Unknown curve type: %v", typ.CurveType)
	}
}

func NewKeyFromPriv(typ KeyType, priv []byte) (*Key, error) {
	switch typ.CurveType {
	case CurveTypeSecp256k1:
		return keyFromPrivSecp256k1(typ.AddrType, priv)
	case CurveTypeEd25519:
		return keyFromPrivEd25519(typ.AddrType, priv)
	default:
		return nil, fmt.Errorf("Unknown curve type: %v", typ.CurveType)
	}
}

func (k *Key) Sign(hash []byte) ([]byte, error) {
	switch k.Type.CurveType {
	case CurveTypeSecp256k1:
		return signSecp256k1(k, hash)
	case CurveTypeEd25519:
		return signEd25519(k, hash)
	}
	return nil, InvalidCurveErr(k.Type.CurveType)
}

func (k *Key) Pubkey() ([]byte, error) {
	switch k.Type.CurveType {
	case CurveTypeSecp256k1:
		return pubKeySecp256k1(k)
	case CurveTypeEd25519:
		return pubKeyEd25519(k)
	}
	return nil, InvalidCurveErr(k.Type.CurveType)
}

func Verify(curveType CurveType, hash, sig, pub []byte) (bool, error) {
	switch curveType {
	case CurveTypeSecp256k1:
		return verifySigSecp256k1(hash, sig, pub)
	case CurveTypeEd25519:
		return verifySigEd25519(hash, sig, pub)
	}
	return false, InvalidCurveErr(curveType)
}

//-----------------------------------------------------------------------------
// json encodings

// addresses should be hex encoded

type plainKeyJSON struct {
	Id         []byte
	Type       string
	Address    string
	PrivateKey []byte
}

type cipherJSON struct {
	Salt       []byte
	Nonce      []byte
	CipherText []byte
}

type encryptedKeyJSON struct {
	Id      []byte
	Type    string
	Address string
	Crypto  cipherJSON
}

func (k *Key) MarshalJSON() (j []byte, err error) {
	jStruct := plainKeyJSON{
		[]byte(k.Id.String()),
		k.Type.String(),
		fmt.Sprintf("%X", k.Address),
		k.PrivateKey,
	}
	j, err = json.Marshal(jStruct)
	return j, err
}

func (k *Key) UnmarshalJSON(j []byte) (err error) {
	keyJSON := new(plainKeyJSON)
	err = json.Unmarshal(j, &keyJSON)
	if err != nil {
		return err
	}
	// TODO: remove this
	if len(keyJSON.PrivateKey) == 0 {
		return NoPrivateKeyErr("")
	}

	u, err := uuid.Parse(string(keyJSON.Id))
	if err != nil {
		return err
	}
	k.Id = u
	k.Address, err = hex.DecodeString(keyJSON.Address)
	if err != nil {
		return err
	}
	k.PrivateKey = keyJSON.PrivateKey
	k.Type, err = KeyTypeFromString(keyJSON.Type)
	return err
}

// returns the address if valid, nil otherwise
func IsValidKeyJson(j []byte) []byte {
	j1 := new(plainKeyJSON)
	e1 := json.Unmarshal(j, &j1)
	if e1 == nil {
		addr, _ := hex.DecodeString(j1.Address)
		return addr
	}

	j2 := new(encryptedKeyJSON)
	e2 := json.Unmarshal(j, &j2)
	if e2 == nil {
		addr, _ := hex.DecodeString(j2.Address)
		return addr
	}

	return nil
}

//-----------------------------------------------------------------------------
// main utility functions for each key type (new, pub, sign, verify)
// TODO: run all sorts of length and validity checks

func newKeySecp256k1(addrType AddrType) *Key {
	pub, priv := secp256k1.GenerateKeyPair()
	id, _ := uuid.NewRandom()
	return &Key{
		Id:         id,
		Type:       KeyType{CurveTypeSecp256k1, addrType},
		Address:    AddressFromPub(addrType, pub),
		PrivateKey: priv,
	}
}

func newKeyEd25519(addrType AddrType) *Key {
	randBytes := randentropy.GetEntropyMixed(32)
	key, _ := keyFromPrivEd25519(addrType, randBytes)
	return key
}

func keyFromPrivSecp256k1(addrType AddrType, priv []byte) (*Key, error) {
	pub, err := secp256k1.GeneratePubKey(priv)
	if err != nil {
		return nil, err
	}
	id, _ := uuid.NewRandom()
	return &Key{
		Id:         id,
		Type:       KeyType{CurveTypeSecp256k1, addrType},
		Address:    AddressFromPub(addrType, pub),
		PrivateKey: priv,
	}, nil
}

func keyFromPrivEd25519(addrType AddrType, priv []byte) (*Key, error) {
	privKeyBytes := new([64]byte)
	copy(privKeyBytes[:32], priv)
	pubKeyBytes := ed25519.MakePublicKey(privKeyBytes)
	pubKey := helpers.PubKeyEd25519(*pubKeyBytes)
	id, _ := uuid.NewRandom()
	return &Key{
		Id:         id,
		Type:       KeyType{CurveTypeEd25519, addrType},
		Address:    pubKey.Address(),
		PrivateKey: privKeyBytes[:],
	}, nil
}

func pubKeySecp256k1(k *Key) ([]byte, error) {
	return secp256k1.GeneratePubKey(k.PrivateKey)
}

func pubKeyEd25519(k *Key) ([]byte, error) {
	priv := k.PrivateKey
	privKeyBytes := new([64]byte)
	copy(privKeyBytes[:32], priv)
	pubKeyBytes := ed25519.MakePublicKey(privKeyBytes)
	return pubKeyBytes[:], nil
}

func signSecp256k1(k *Key, hash []byte) ([]byte, error) {
	return secp256k1.Sign(hash, k.PrivateKey)
}

func signEd25519(k *Key, hash []byte) ([]byte, error) {
	priv := k.PrivateKey
	var privKey helpers.PrivKeyEd25519
	copy(privKey[:], priv)
	sig := privKey.Sign(hash)
	sigB := sig.(helpers.SignatureEd25519)
	return sigB[:], nil
}

func verifySigSecp256k1(hash, sig, pubOG []byte) (bool, error) {
	pub, err := secp256k1.RecoverPubkey(hash, sig)
	if err != nil {
		return false, err
	}

	if bytes.Compare(pub, pubOG) != 0 {
		return false, fmt.Errorf("Recovered pub key does not match. Got %X, expected %X", pub, pubOG)
	}

	// TODO: validate recovered pub!

	return true, nil
}

func verifySigEd25519(hash, sig, pub []byte) (bool, error) {
	pubKeyBytes := new([32]byte)
	copy(pubKeyBytes[:], pub)
	sigBytes := new([64]byte)
	copy(sigBytes[:], sig)
	res := ed25519.Verify(pubKeyBytes, hash, sigBytes)
	return res, nil
}
