package blockchain

import (
	"testing"

	"github.com/hyperledger/burrow/genesis"
	"github.com/stretchr/testify/assert"
	dbm "github.com/tendermint/tmlibs/db"
)

func TestSerialise(t *testing.T) {
	testGenesisDoc, _, _ := genesis.NewDeterministicGenesis(123).GenesisDoc(1, true, 1000, 1, true, 1000)
	testDB := dbm.NewDB("test", dbm.MemDBBackend, ".")
	bc1 := newBlockchain(testDB, testGenesisDoc)
	bc1.validatorSet = newValidatorSet(testGenesisDoc.GetMaximumPower(), testGenesisDoc.Validators())
	bc1.save()
	bc2, err := loadBlockchain(testDB)

	assert.NoError(t, err)
	assert.Equal(t, bc1, bc2)
}
