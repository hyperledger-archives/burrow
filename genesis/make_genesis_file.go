package genesis

import (
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/eris-ltd/common/go/common"
	stypes "github.com/eris-ltd/eris-db/manager/eris-mint/state/types"
	ptypes "github.com/eris-ltd/eris-db/permission/types"

	"github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
)

//------------------------------------------------------------------------------------
// core functions

func GenerateKnown(chainID, accountsPathCSV, validatorsPathCSV string) ([]byte, error) {
	var genDoc *stypes.GenesisDoc
	var err error

	// TODO [eb] eliminate reading priv_val ... [zr] where?
	if accountsPathCSV == "" || validatorsPathCSV == "" {
		return []byte{}, fmt.Errorf("both accounts.csv and validators.csv is required")

	}

	pubkeys, amts, names, perms, setbits, err := parseCsv(validatorsPathCSV)
	if err != nil {
		return nil, err
	}

	pubkeysA, amtsA, namesA, permsA, setbitsA, err := parseCsv(accountsPathCSV)
	if err != nil {
		return nil, err
	}

	genDoc = newGenDoc(chainID, len(pubkeys), len(pubkeysA))
	for i, pk := range pubkeys {
		genDocAddValidator(genDoc, pk, amts[i], names[i], perms[i], setbits[i], i)
	}
	for i, pk := range pubkeysA {
		genDocAddAccount(genDoc, pk, amtsA[i], namesA[i], permsA[i], setbitsA[i], i)
	}

	buf, buf2, n := new(bytes.Buffer), new(bytes.Buffer), new(int)
	wire.WriteJSON(genDoc, buf, n, &err)
	if err != nil {
		return nil, err
	}
	if err := json.Indent(buf2, buf.Bytes(), "", "\t"); err != nil {
		return nil, err
	}
	genesisBytes := buf2.Bytes()

	return genesisBytes, nil
}

//-----------------------------------------------------------------------------
// gendoc convenience functions

func newGenDoc(chainID string, nVal, nAcc int) *stypes.GenesisDoc {
	genDoc := stypes.GenesisDoc{
		ChainID: chainID,
		// GenesisTime: time.Now(),
	}
	genDoc.Accounts = make([]stypes.GenesisAccount, nAcc)
	genDoc.Validators = make([]stypes.GenesisValidator, nVal)
	return &genDoc
}

func genDocAddAccount(genDoc *stypes.GenesisDoc, pubKey crypto.PubKeyEd25519, amt int64, name string, perm, setbit ptypes.PermFlag, index int) {
	addr := pubKey.Address()
	acc := stypes.GenesisAccount{
		Address: addr,
		Amount:  amt,
		Name:    name,
		Permissions: &ptypes.AccountPermissions{
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

func genDocAddValidator(genDoc *stypes.GenesisDoc, pubKey crypto.PubKeyEd25519, amt int64, name string, perm, setbit ptypes.PermFlag, index int) {
	addr := pubKey.Address()
	genDoc.Validators[index] = stypes.GenesisValidator{
		PubKey: pubKey,
		Amount: amt,
		Name:   name,
		UnbondTo: []stypes.BasicAccount{
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
func parseCsv(filePath string) (pubKeys []crypto.PubKeyEd25519, amts []int64, names []string, perms, setbits []ptypes.PermFlag, err error) {

	csvFile, err := os.Open(filePath)
	if err != nil {
		common.Exit(fmt.Errorf("Couldn't open file: %s: %v", filePath, err))
	}
	defer csvFile.Close()

	r := csv.NewReader(csvFile)
	//r.FieldsPerRecord = # of records expected
	params, err := r.ReadAll()
	if err != nil {
		common.Exit(fmt.Errorf("Couldn't read file: %v", err))

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
		permsS[i] = ifExistsElse(each, 3, fmt.Sprintf("%d", ptypes.DefaultPermFlags))
		setbitS[i] = ifExistsElse(each, 4, permsS[i])
	}

	//TODO convert int to uint64, see issue #25
	perms = make([]ptypes.PermFlag, len(permsS))
	for i, perm := range permsS {
		pflag, err := strconv.Atoi(perm)
		if err != nil {
			common.Exit(fmt.Errorf("Permissions must be an integer"))
		}
		perms[i] = ptypes.PermFlag(pflag)
	}
	setbits = make([]ptypes.PermFlag, len(setbitS))
	for i, setbit := range setbitS {
		setbitsFlag, err := strconv.Atoi(setbit)
		if err != nil {
			common.Exit(fmt.Errorf("SetBits must be an integer"))
		}
		setbits[i] = ptypes.PermFlag(setbitsFlag)
	}

	// convert amts to ints
	amts = make([]int64, len(amtS))
	for i, a := range amtS {
		if amts[i], err = strconv.ParseInt(a, 10, 64); err != nil {
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
