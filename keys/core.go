package keys

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	tmint_crypto "github.com/hyperledger/burrow/crypto/helpers"
	wire "github.com/tendermint/go-wire"
)

const (
	DefaultHost     = "localhost"
	DefaultPort     = "10997"
	DefaultHashType = "sha256"
	DefaultKeysDir  = ".keys"
	TestPort        = "7674"
)

func returnDataDir(dir string) (string, error) {
	dir = path.Join(dir, "data")
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return dir, checkMakeDataDir(dir)
}

func returnNamesDir(dir string) (string, error) {
	dir = path.Join(dir, "names")
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return dir, checkMakeDataDir(dir)
}

//-----

func NewKeyStore(dir string) KeyStore {
	return KeyStore{keysDirPath: dir}
}

//----------------------------------------------------------------
func writeKey(keyDir string, addr, keyJson []byte) ([]byte, error) {
	dir, err := returnDataDir(keyDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to get keys dir: %v", err)
	}
	if err := WriteKeyFile(addr, dir, keyJson); err != nil {
		return nil, err
	}
	return addr, nil
}

func coreExport(key *Key) ([]byte, error) {
	type privValidator struct {
		Address    []byte        `json:"address"`
		PubKey     []interface{} `json:"pub_key"`
		PrivKey    []interface{} `json:"priv_key"`
		LastHeight int           `json:"last_height"`
		LastRound  int           `json:"last_round"`
		LastStep   int           `json:"last_step"`
	}

	pub := key.Pubkey()

	var pubKeyWithType []interface{}
	var pubKey tmint_crypto.PubKeyEd25519
	copy(pubKey[:], pub)
	pubKeyWithType = append(pubKeyWithType, tmint_crypto.PubKeyTypeEd25519)
	pubKeyWithType = append(pubKeyWithType, pubKey)

	var privKeyWithType []interface{}
	var privKey tmint_crypto.PrivKeyEd25519
	copy(privKey[:], key.PrivateKey.RawBytes())
	privKeyWithType = append(privKeyWithType, tmint_crypto.PrivKeyTypeEd25519)
	privKeyWithType = append(privKeyWithType, privKey)

	privVal := &privValidator{
		Address: key.Address[:],
		PubKey:  pubKeyWithType,
		PrivKey: privKeyWithType,
	}

	return wire.JSONBytes(privVal), nil
}

//----------------------------------------------------------------
// manage names for keys

func coreNameAdd(keysDir, name, addr string) error {
	namesDir, err := returnNamesDir(keysDir)
	if err != nil {
		return err
	}
	dataDir, err := returnDataDir(keysDir)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path.Join(dataDir, addr+".json")); err != nil {
		return fmt.Errorf("Unknown key %s", addr)
	}
	return ioutil.WriteFile(path.Join(namesDir, name), []byte(addr), 0600)
}

func coreNameList(keysDir string) (map[string]string, error) {
	dir, err := returnNamesDir(keysDir)
	if err != nil {
		return nil, err
	}
	names := make(map[string]string)
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, f := range fs {
		b, err := ioutil.ReadFile(path.Join(dir, f.Name()))
		if err != nil {
			return nil, err
		}
		names[f.Name()] = string(b)
	}
	return names, nil
}

func coreAddrList(keysDir string) (map[int]string, error) {
	dir, err := returnDataDir(keysDir)
	if err != nil {
		return nil, err
	}
	addrs := make(map[int]string)
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(fs); i++ {
		addrs[i] = fs[i].Name()
	}
	return addrs, nil
}

func coreNameRm(keysDir string, name string) error {
	dir, err := returnNamesDir(keysDir)
	if err != nil {
		return err
	}
	return os.Remove(path.Join(dir, name))
}

func coreNameGet(keysDir, name string) (string, error) {
	dir, err := returnNamesDir(keysDir)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadFile(path.Join(dir, name))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func checkMakeDataDir(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return err
		}
	}
	return nil
}

// return addr from name or addr
func getNameAddr(keysDir, name, addr string) (string, error) {
	if name == "" && addr == "" {
		return "", fmt.Errorf("at least one of name or addr must be provided")
	}

	// name takes precedent if both are given
	var err error
	if name != "" {
		addr, err = coreNameGet(keysDir, name)
		if err != nil {
			return "", err
		}
	}
	return strings.ToUpper(addr), nil
}
