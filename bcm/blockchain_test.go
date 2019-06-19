package bcm

import (
	"testing"
	"time"

	"github.com/hyperledger/burrow/crypto/sha3"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestLoadOrNewBlockchain(t *testing.T) {
	// Initial load - fresh chain
	genesisDoc := newGenesisDoc()
	db := dbm.NewMemDB()
	blockchain, exists, err := LoadOrNewBlockchain(db, genesisDoc, logging.NewNoopLogger())
	require.NoError(t, err)
	assert.False(t, exists)
	assert.Equal(t, genesisDoc.GenesisTime, blockchain.LastBlockTime())
	assert.Equal(t, uint64(0), blockchain.LastBlockHeight())
	assert.Equal(t, genesisDoc.Hash(), blockchain.AppHashAfterLastBlock())

	// First block
	blockTime1 := genesisDoc.GenesisTime.Add(time.Second * 10)
	blockHash1 := sha3.Sha3([]byte("blockHash"))
	appHash1 := sha3.Sha3([]byte("appHash"))
	err = blockchain.CommitBlock(blockTime1, blockHash1, appHash1)
	require.NoError(t, err)
	assertState(t, blockchain, 1, blockTime1, appHash1)

	// Second block
	blockTime2a := blockTime1.Add(time.Second * 30)
	blockHash2a := sha3.Sha3(append(blockHash1, 2))
	appHash2a := sha3.Sha3(append(appHash1, 2))
	err = blockchain.CommitBlock(blockTime2a, blockHash2a, appHash2a)
	require.NoError(t, err)
	assertState(t, blockchain, 2, blockTime2a, appHash2a)

	// Load at checkpoint (i.e. first block)
	blockchain, exists, err = LoadOrNewBlockchain(db, genesisDoc, logging.NewNoopLogger())
	require.NoError(t, err)
	assert.True(t, exists)
	// Assert first block values
	assertState(t, blockchain, 1, blockTime1, appHash1)

	// Commit (overwriting previous block 2 pointer
	blockTime2b := blockTime1.Add(time.Second * 30)
	blockHash2b := sha3.Sha3(append(blockHash1, 2))
	appHash2b := sha3.Sha3(append(appHash1, 2))
	err = blockchain.CommitBlock(blockTime2b, blockHash2b, appHash2b)
	require.NoError(t, err)
	assertState(t, blockchain, 2, blockTime2b, appHash2b)

	// Commit again to check things are okay
	blockTime3 := blockTime2b.Add(time.Second * 30)
	blockHash3 := sha3.Sha3(append(blockHash2b, 2))
	appHash3 := sha3.Sha3(append(appHash2b, 2))
	err = blockchain.CommitBlock(blockTime3, blockHash3, appHash3)
	require.NoError(t, err)
	assertState(t, blockchain, 3, blockTime3, appHash3)

	// Load at checkpoint (i.e. block 2b)
	blockchain, exists, err = LoadOrNewBlockchain(db, genesisDoc, logging.NewNoopLogger())
	require.NoError(t, err)
	assert.True(t, exists)
	// Assert first block values
	assertState(t, blockchain, 2, blockTime2b, appHash2b)
}

func assertState(t *testing.T, blockchain *Blockchain, height uint64, blockTime time.Time, appHash []byte) {
	assert.Equal(t, height, blockchain.LastBlockHeight())
	assert.Equal(t, blockTime, blockchain.LastBlockTime())
	assert.Equal(t, appHash, blockchain.AppHashAfterLastBlock())
}

func newGenesisDoc() *genesis.GenesisDoc {
	genesisDoc, _, _ := genesis.NewDeterministicGenesis(3450976).GenesisDoc(23, 10)
	return genesisDoc
}
