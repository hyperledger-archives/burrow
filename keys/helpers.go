package keys

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hyperledger/burrow/acm"
)

var (
	scratchDir = "test_scratch"
)

func EnterTestKeyStore(privateAccounts ...*acm.PrivateAccount) (keystore *KeyStore, cleanup func()) {
	testDir, err := ioutil.TempDir("", scratchDir)
	if err != nil {
		panic(fmt.Errorf("could not make temp dir for integration tests: %v", err))
	}
	// If you need to inspectdirs
	//testDir := scratchDir
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	ks := NewKeyStore(testDir, true)
	ks.AddPrivateAccounts(privateAccounts...)
	return ks, func() { os.RemoveAll(testDir) }
}

func (ks *KeyStore) AddPrivateAccounts(privateAccounts ...*acm.PrivateAccount) {
	for _, pa := range privateAccounts {
		key := &Key{
			Address:    pa.GetAddress(),
			PublicKey:  pa.GetPublicKey(),
			PrivateKey: pa.PrivateKey(),
			CurveType:  pa.PrivateKey().CurveType,
		}
		ks.StoreKey("", key)
	}
}
