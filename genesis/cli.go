package genesis

import (
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	. "github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/mint-client/Godeps/_workspace/src/github.com/eris-ltd/tendermint/types"
	//"github.com/spf13/cobra"
	wire "github.com/tendermint/go-wire"

	"github.com/eris-ltd/mint-client/Godeps/_workspace/src/github.com/eris-ltd/tendermint/account"
	ptypes "github.com/eris-ltd/mint-client/Godeps/_workspace/src/github.com/eris-ltd/tendermint/permission/types"
	stypes "github.com/eris-ltd/mint-client/Godeps/_workspace/src/github.com/eris-ltd/tendermint/state/types"
)

//------------------------------------------------------------------------------------
// core functions

func GenerateKnown(chainID, csvFile, pubKeys string) ([]byte, error) {
	var genDoc *stypes.GenesisDoc
	var err error
	// either we pass the name of a csv file or we read a priv_validator over stdin
	// TODO eliminate reading priv_val
	if csvFile != "" {
		var csvValidators, csvAccounts string
		csvFiles := strings.Split(csvFile, ",")
		csvValidators = csvFiles[0]
		if len(csvFiles) > 1 {
			csvAccounts = csvFiles[1]
		}
		pubkeys, amts, names, perms, setbits, err := parseCsv(csvValidators)
		if err != nil {
			return nil, err
		}

		if csvAccounts == "" {
			genDoc = newGenDoc(chainID, len(pubkeys), len(pubkeys))
			for i, pk := range pubkeys {
				genDocAddAccountAndValidator(genDoc, pk, amts[i], names[i], perms[i], setbits[i], i)
			}
		} else {
			pubkeysA, amtsA, namesA, permsA, setbitsA, err := parseCsv(csvAccounts)
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
		}
	} else if pubKeys != "" {
		pubkeys := strings.Split(pubKeys, ",")
		amt := int64(1) << 50
		pubKeys := pubKeyStringsToPubKeys(pubkeys)
		genDoc = newGenDoc(chainID, len(pubkeys), len(pubkeys))

		for i, pk := range pubKeys {
			genDocAddAccountAndValidator(genDoc, pk, amt, "", ptypes.DefaultPermFlags, ptypes.DefaultPermFlags, i)
		}
	} else {
		privJSON := readStdinTimeout()
		genDoc = genesisFromPrivValBytes(chainID, privJSON)
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

// genesis file with only one validator, using priv_validator.json
func genesisFromPrivValBytes(chainID string, privJSON []byte) *stypes.GenesisDoc {
	var err error
	privVal := wire.ReadJSON(&types.PrivValidator{}, privJSON, &err).(*types.PrivValidator)
	if err != nil {
		Exit(fmt.Errorf("Error reading PrivValidator on stdin: %v\n", err))
	}
	pubKey := privVal.PubKey
	amt := int64(1) << 50

	genDoc := newGenDoc(chainID, 1, 1)

	genDocAddAccountAndValidator(genDoc, pubKey, amt, "", ptypes.DefaultPermFlags, ptypes.DefaultPermFlags, 0)

	return genDoc
}

func genDocAddAccount(genDoc *stypes.GenesisDoc, pubKey account.PubKeyEd25519, amt int64, name string, perm, setbit ptypes.PermFlag, index int) {
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

func genDocAddValidator(genDoc *stypes.GenesisDoc, pubKey account.PubKeyEd25519, amt int64, name string, perm, setbit ptypes.PermFlag, index int) {
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
}

func genDocAddAccountAndValidator(genDoc *stypes.GenesisDoc, pubKey account.PubKeyEd25519, amt int64, name string, perm, setbit ptypes.PermFlag, index int) {
	genDocAddAccount(genDoc, pubKey, amt, name, perm, setbit, index)
	genDocAddValidator(genDoc, pubKey, amt, name, perm, setbit, index)
}

//-----------------------------------------------------------------------------
// util functions

// convert hex strings to ed25519 pubkeys
func pubKeyStringsToPubKeys(pubkeys []string) []account.PubKeyEd25519 {
	pubKeys := make([]account.PubKeyEd25519, len(pubkeys))
	for i, k := range pubkeys {
		pubBytes, err := hex.DecodeString(k)
		if err != nil {
			Exit(fmt.Errorf("Pubkey (%s) is invalid hex: %v", k, err))
		}
		copy(pubKeys[i][:], pubBytes)
	}
	return pubKeys
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

//takes a csv in the format defined [here]
func parseCsv(filePath string) (pubKeys []account.PubKeyEd25519, amts []int64, names []string, perms, setbits []ptypes.PermFlag, err error) {

	csvFile, err := os.Open(filePath)
	if err != nil {
		Exit(fmt.Errorf("Couldn't open file: %s: %v", filePath, err))
	}
	defer csvFile.Close()

	r := csv.NewReader(csvFile)
	//r.FieldsPerRecord = # of records expected
	params, err := r.ReadAll()
	if err != nil {
		Exit(fmt.Errorf("Couldn't read file: %v", err))

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
			Exit(fmt.Errorf("Permissions must be an integer"))
		}
		perms[i] = ptypes.PermFlag(pflag)
	}
	setbits = make([]ptypes.PermFlag, len(setbitS))
	for i, setbit := range setbitS {
		setbitsFlag, err := strconv.Atoi(setbit)
		if err != nil {
			Exit(fmt.Errorf("SetBits must be an integer"))
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
	pubKeys = pubKeyStringsToPubKeys(pubkeys)

	return pubKeys, amts, names, perms, setbits, nil
}

const stdinTimeoutSeconds = 1

// read the priv validator json off stdin or timeout and fail
func readStdinTimeout() []byte {
	ch := make(chan []byte, 1)
	go func() {
		privJSON, err := ioutil.ReadAll(os.Stdin)
		IfExit(err)
		ch <- privJSON
	}()
	ticker := time.Tick(time.Second * stdinTimeoutSeconds)
	select {
	case <-ticker:
		Exit(fmt.Errorf("Please pass a priv_validator.json on stdin, or specify either a pubkey with --pub or csv file with --csv"))
	case privJSON := <-ch:
		return privJSON
	}
	return nil
}
