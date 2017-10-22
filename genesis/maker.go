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

package genesis

import (
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/permission"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/hyperledger/burrow/util"
	"github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
)

const (
	PublicKeyEd25519ByteLength   int = 32
	PublicKeySecp256k1ByteLength int = 64
)

// NewGenesisAccount returns a new GenesisAccount
func NewGenesisAccount(address acm.Address, amount uint64, name string,
	permissions *ptypes.AccountPermissions) *GenesisAccount {
	return &GenesisAccount{
		BasicAccount: BasicAccount{
			Address: address,
			Amount:  amount,
		},
		Name:        name,
		Permissions: permissions.Clone(),
	}
}

func NewGenesisValidator(amount uint64, name string, unbondToAddress acm.Address,
	unbondAmount uint64, keyType string, publicKeyBytes []byte) (*GenesisValidator, error) {
	// convert the key bytes into a typed fixed size byte array
	var typedPublicKeyBytes []byte
	switch keyType {
	case "ed25519": // ed25519 has type byte 0x01
		// TODO: [ben] functionality and checks need to be inherit in the type
		if len(publicKeyBytes) != PublicKeyEd25519ByteLength {
			return nil, fmt.Errorf("Invalid length provided for ed25519 public key (%v bytes provided but expected %v bytes)",
				len(publicKeyBytes), PublicKeyEd25519ByteLength)
		}
		// prepend type byte to public key
		typedPublicKeyBytes = append([]byte{crypto.TypeEd25519}, publicKeyBytes...)
	case "secp256k1": // secp256k1 has type byte 0x02
		if len(publicKeyBytes) != PublicKeySecp256k1ByteLength {
			return nil, fmt.Errorf("Invalid length provided for secp256k1 public key (%v bytes provided but expected %v bytes)",
				len(publicKeyBytes), PublicKeySecp256k1ByteLength)
		}
		// prepend type byte to public key
		typedPublicKeyBytes = append([]byte{crypto.TypeSecp256k1}, publicKeyBytes...)
	default:
		return nil, fmt.Errorf("Unsupported key type (%s)", keyType)
	}
	newPublicKey, err := crypto.PubKeyFromBytes(typedPublicKeyBytes)
	if err != nil {
		return nil, err
	}
	// ability to unbond to multiple accounts currently unused
	var unbondTo []BasicAccount

	return &GenesisValidator{
		PubKey: newPublicKey,
		Amount: unbondAmount,
		Name:   name,
		UnbondTo: append(unbondTo, BasicAccount{
			Address: unbondToAddress,
			Amount:  unbondAmount,
		}),
	}, nil
}

//------------------------------------------------------------------------------------
// interface functions that are consumed by monax tooling

func GenerateKnown(chainName, accountsPathCSV, validatorsPathCSV string) (string, error) {
	return generateKnownWithTime(chainName, accountsPathCSV, validatorsPathCSV,
		// set the timestamp for the genesis
		time.Now())
}

//------------------------------------------------------------------------------------
// interface functions that are consumed by monax tooling

func GenerateGenesisFileBytes(chainName string, salt []byte, genesisTime time.Time, genesisAccounts map[string]acm.Account,
	genesisValidators map[string]acm.Validator) ([]byte, error) {

	genesisDoc := MakeGenesisDocFromAccounts(chainName, salt, genesisTime, genesisAccounts, genesisValidators)

	var err error
	buf, buf2, n := new(bytes.Buffer), new(bytes.Buffer), new(int)
	wire.WriteJSON(genesisDoc, buf, n, &err)
	if err != nil {
		return nil, err
	}
	if err := json.Indent(buf2, buf.Bytes(), "", "\t"); err != nil {
		return nil, err
	}

	return buf2.Bytes(), nil
}

//------------------------------------------------------------------------------------
// core functions that provide functionality for monax tooling in v0.16

// GenerateKnownWithTime takes chainName, an accounts and validators CSV filepath
// and a timestamp to generate the string of `genesis.json`
// NOTE: [ben] is introduced as technical debt to preserve the signature
// of GenerateKnown but in order to introduce the timestamp gradually
// This will be deprecated in v0.17
func generateKnownWithTime(chainName, accountsPathCSV, validatorsPathCSV string,
	genesisTime time.Time) (string, error) {
	var genDoc *GenesisDoc

	// TODO [eb] eliminate reading priv_val ... [zr] where?
	if accountsPathCSV == "" || validatorsPathCSV == "" {
		return "", fmt.Errorf("both accounts.csv and validators.csv is required")
	}

	pubkeys, amts, names, perms, setbits, err := parseCsv(validatorsPathCSV)
	if err != nil {
		return "", err
	}

	pubkeysA, amtsA, namesA, permsA, setbitsA, err := parseCsv(accountsPathCSV)
	if err != nil {
		return "", err
	}

	genDoc = newGenDoc(chainName, genesisTime, len(pubkeys), len(pubkeysA))
	for i, pk := range pubkeys {
		genDocAddValidator(genDoc, pk, amts[i], names[i], perms[i], setbits[i], i)
	}
	for i, pk := range pubkeysA {
		genDocAddAccount(genDoc, pk, amtsA[i], namesA[i], permsA[i], setbitsA[i], i)
	}

	buf, buf2, n := new(bytes.Buffer), new(bytes.Buffer), new(int)
	wire.WriteJSON(genDoc, buf, n, &err)
	if err != nil {
		return "", err
	}
	if err := json.Indent(buf2, buf.Bytes(), "", "\t"); err != nil {
		return "", err
	}

	return buf2.String(), nil
}

//-----------------------------------------------------------------------------
// gendoc convenience functions

func newGenDoc(chainName string, genesisTime time.Time, nVal, nAcc int) *GenesisDoc {
	genDoc := GenesisDoc{
		ChainName:   chainName,
		GenesisTime: genesisTime,
	}
	genDoc.Accounts = make([]GenesisAccount, nAcc)
	genDoc.Validators = make([]GenesisValidator, nVal)
	return &genDoc
}

func genDocAddAccount(genDoc *GenesisDoc, pubKey crypto.PubKeyEd25519, amt uint64, name string,
	perm, setbit ptypes.PermFlag, index int) {
	addr, _ := acm.AddressFromBytes(pubKey.Address())
	acc := GenesisAccount{
		BasicAccount: BasicAccount{
			Address: addr,
			Amount:  amt,
		},
		Name: name,
		Permissions: ptypes.AccountPermissions{
			Base: ptypes.BasePermissions{
				Perms:  perm,
				SetBit: setbit,
			},
		},
	}
	if index < 0 {
		genDoc.Accounts = append(genDoc.Accounts, acc)
	} else {
		genDoc.Accounts[index] = acc
	}
}

func genDocAddValidator(genDoc *GenesisDoc, pubKey crypto.PubKeyEd25519, amt uint64, name string,
	perm, setbit ptypes.PermFlag, index int) {
	addr, _ := acm.AddressFromBytes(pubKey.Address())
	genDoc.Validators[index] = GenesisValidator{
		PubKey: pubKey.Wrap(),
		Amount: amt,
		Name:   name,
		UnbondTo: []BasicAccount{
			{
				Address: addr,
				Amount:  amt,
			},
		},
	}
	// [zr] why no index < 0 like in genDocAddAccount?
}

//-----------------------------------------------------------------------------
// util functions

// convert hex strings to ed25519 pubkeys
func pubKeyStringsToPubKeys(pubkeys []string) ([]crypto.PubKeyEd25519, error) {
	pubKeys := make([]crypto.PubKeyEd25519, len(pubkeys))
	for i, k := range pubkeys {
		pubBytes, err := hex.DecodeString(k)
		if err != nil {
			return pubKeys, err
		}
		copy(pubKeys[i][:], pubBytes)
	}
	return pubKeys, nil
}

// empty is over written
func ifExistsElse(list []string, index int, defaultValue string) string {
	if len(list) > index {
		if list[index] != "" {
			return list[index]
		}
	}
	return defaultValue
}

// takes a csv in the following format: pubkey, starting balance, name, permissions, setbit
func parseCsv(filePath string) (pubKeys []crypto.PubKeyEd25519, amts []uint64, names []string, perms, setbits []ptypes.PermFlag, err error) {

	csvFile, err := os.Open(filePath)
	if err != nil {
		util.Fatalf("Couldn't open file: %s: %v", filePath, err)
	}
	defer csvFile.Close()

	r := csv.NewReader(csvFile)
	//r.FieldsPerRecord = # of records expected
	params, err := r.ReadAll()
	if err != nil {
		util.Fatalf("Couldn't read file: %v", err)

	}

	pubkeys := make([]string, len(params))
	amtS := make([]string, len(params))
	names = make([]string, len(params))
	permsS := make([]string, len(params))
	setbitS := make([]string, len(params))
	for i, each := range params {
		pubkeys[i] = each[0]
		amtS[i] = ifExistsElse(each, 1, "1000")
		names[i] = ifExistsElse(each, 2, "")
		permsS[i] = ifExistsElse(each, 3, fmt.Sprintf("%d", permission.DefaultPermFlags))
		setbitS[i] = ifExistsElse(each, 4, permsS[i])
	}

	//TODO convert int to uint64, see issue #25
	perms = make([]ptypes.PermFlag, len(permsS))
	for i, perm := range permsS {
		pflag, err := strconv.Atoi(perm)
		if err != nil {
			util.Fatalf("Permissions (%v) must be an integer", perm)
		}
		perms[i] = ptypes.PermFlag(pflag)
	}
	setbits = make([]ptypes.PermFlag, len(setbitS))
	for i, setbit := range setbitS {
		setbitsFlag, err := strconv.Atoi(setbit)
		if err != nil {
			util.Fatalf("SetBits (%v) must be an integer", setbit)
		}
		setbits[i] = ptypes.PermFlag(setbitsFlag)
	}

	// convert amts to ints
	amts = make([]uint64, len(amtS))
	for i, a := range amtS {
		if amts[i], err = strconv.ParseUint(a, 10, 64); err != nil {
			err = fmt.Errorf("Invalid amount: %v", err)
			return
		}
	}

	// convert pubkey hex strings to struct
	pubKeys, err = pubKeyStringsToPubKeys(pubkeys)
	if err != nil {
		return
	}

	return pubKeys, amts, names, perms, setbits, nil
}
