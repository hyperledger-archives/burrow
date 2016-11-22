package genesis

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	ptypes "github.com/eris-ltd/mint-client/Godeps/_workspace/src/github.com/eris-ltd/tendermint/permission/types"
	stypes "github.com/eris-ltd/mint-client/Godeps/_workspace/src/github.com/eris-ltd/tendermint/state/types"
	"github.com/eris-ltd/mint-client/Godeps/_workspace/src/github.com/eris-ltd/tendermint/types"
)

func MakeGenesisDocFromFile(genDocFile string) *stypes.GenesisDoc {
	jsonBlob, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		fmt.Sprintf("Couldn't read GenesisDoc file: %v", err)
		os.Exit(1)
	}
	return stypes.GenesisDocFromJSON(jsonBlob)
}

func testCoreRandom(N int) error {
	chainID := "test_chainID"

	genBytes, privVals, err := coreRandom(N, chainID, "", "", "", false)
	if err != nil {
		return err
	}

	if len(privVals) != N {
		return fmt.Errorf("len(privVals) != N")
	}

	// make sure each validator is in the genesis and all genesi are the same
	for i, v := range privVals {
		dirFlag := DirFlag
		if N > 1 {
			dirFlag = path.Join(DirFlag, fmt.Sprintf("%s_%d", chainID, i))
		}

		b, err := ioutil.ReadFile(path.Join(dirFlag, "genesis.json"))
		if err != nil {
			return err
		}
		if !bytes.Equal(b, genBytes) {
			return fmt.Errorf("written genesis.json different from returned by coreRandom")
		}

		gDoc := MakeGenesisDocFromFile(path.Join(dirFlag, "genesis.json"))

		if len(gDoc.Validators) != N {
			return fmt.Errorf("Expected %d validators. Got %d", N, len(gDoc.Validators))
		}

		privVal := types.LoadPrivValidator(path.Join(dirFlag, "priv_validator.json"))
		if !bytes.Equal(privVal.Address, v.Address) {
			return fmt.Errorf("priv_validator file contents different than result of coreRandom")
		}
		var found bool
		for _, val := range gDoc.Validators {
			if bytes.Equal(val.UnbondTo[0].Address, privVal.Address) {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("failed to find validator %d:%X in genesis.json", i, v.Address)
		}
	}
	return nil
}

func TestRandom(t *testing.T) {
	// make temp dir
	dir, err := ioutil.TempDir(os.TempDir(), "mintgen-test")
	if err != nil {
		t.Fatal(err)
	}

	DirFlag = dir
	defer func() {
		// cleanup
		os.RemoveAll(DirFlag)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if err = testCoreRandom(1); err != nil {
		return
	}
	if err = testCoreRandom(3); err != nil {
		return
	}
}

type GenDoc struct {
	pubkeys []string
	amts    []int
	names   []string
	perms   []int
	setbits []int
}

var csv1 = GenDoc{
	pubkeys: []string{"3D64963C2EE465AA3866DAF420FA1D35F54A1C2DDCF4524C587CD7295D961C09", "3D64963C2EE465AA3866DAF420FA1D35F54A1C2DDCF4524C587CD7295D961C10", "3D64963C2EE465AA3866DAF420FA1D35F54A1C2DDCF4524C587CD7295D961Cff", "3D64963C2EE465AA3866DAF420FA1D35F54A1C2DDCF4524C587CD7295D961Cab"},
	amts:    []int{10, 100, 1000, 100000},
	names:   []string{"", "ok", "hi", "hm"},
	perms:   []int{1, 2, 128, 130},
	setbits: []int{1, 2, 128, 131},
}

var csv2 = GenDoc{
	pubkeys: []string{"3D64963C2EE465AA3866DAF420FA1D35F54A1C2DDCF4524C587CD7295D961C09", "3D64963C2EE465AA3866DAF420FA1D35F54A1C2DDCF4524C587CD7295D961C10", "3D64963C2EE465AA3866DAF420FA1D35F54A1C2DDCF4524C587CD7295D961Cff", "3D64963C2EE465AA3866DAF420FA1D35F54A1C2DDCF4524C587CD7295D961Cab"},
	amts:    nil,
	names:   nil,
	perms:   nil,
	setbits: nil,
}

func csv1String() string {
	buf := new(bytes.Buffer)
	for i, pub := range csv1.pubkeys {
		buf.WriteString(fmt.Sprintf("%s,%d,%s,%d,%d\n", pub, csv1.amts[i], csv1.names[i], csv1.perms[i], csv1.setbits[i]))
	}
	return string(buf.Bytes())
}

func csv2String() string {
	buf := new(bytes.Buffer)
	for _, pub := range csv2.pubkeys {
		buf.WriteString(fmt.Sprintf("%s,\n", pub))
	}
	return string(buf.Bytes())
}

func testKnownCSV(csvFile string, csv GenDoc) error {
	chainID := "test_chainID"

	if err := ioutil.WriteFile(path.Join(DirFlag, "accounts.csv"), []byte(csvFile), 0600); err != nil {
		return err
	}

	genBytes, err := coreKnown(chainID, path.Join(DirFlag, "accounts.csv"), "")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path.Join(DirFlag, "genesis.json"), genBytes, 0600); err != nil {
		return err
	}

	gDoc := MakeGenesisDocFromFile(path.Join(DirFlag, "genesis.json"))

	N := len(csv.pubkeys)
	if len(gDoc.Validators) != N {
		return fmt.Errorf("Expected %d validators. Got %d", N, len(gDoc.Validators))
	}

	for i, pub := range csv.pubkeys {
		pubBytes, _ := hex.DecodeString(pub)
		if !bytes.Equal(gDoc.Validators[i].PubKey[:], pubBytes) {
			return fmt.Errorf("failed to find validator %d:%X in genesis.json", i, pub)
		}
		if len(csv.amts) > 0 && gDoc.Accounts[i].Amount != int64(csv.amts[i]) {
			return fmt.Errorf("amts dont match. got %d, expected %d", gDoc.Accounts[i].Amount, csv.amts[i])
		}
		if len(csv.perms) > 0 && gDoc.Accounts[i].Permissions.Base.Perms != ptypes.PermFlag(csv.perms[i]) {
			return fmt.Errorf("perms dont match. got %d, expected %d", gDoc.Accounts[i].Permissions.Base.Perms, csv.perms[i])
		}
		if len(csv.setbits) > 0 && gDoc.Accounts[i].Permissions.Base.SetBit != ptypes.PermFlag(csv.setbits[i]) {
			return fmt.Errorf("setbits dont match. got %d, expected %d", gDoc.Accounts[i].Permissions.Base.SetBit, csv.setbits[i])
		}
	}
	return nil
}

func TestKnownCSV(t *testing.T) {
	// make temp dir
	dir, err := ioutil.TempDir(os.TempDir(), "mintgen-test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		//cleanup
		os.RemoveAll(DirFlag)
		if err != nil {
			t.Fatal(err)
		}

	}()

	DirFlag = dir
	if err = testKnownCSV(csv1String(), csv1); err != nil {
		return
	}
	if err = testKnownCSV(csv2String(), csv2); err != nil {
		return
	}
}
