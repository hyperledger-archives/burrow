package blockchain

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/db"
)

func TestBlockchain_Encode(t *testing.T) {
	genesisDoc, _, validators := genesis.NewDeterministicGenesis(234).
		GenesisDoc(5, true, 232, 3, true, 34)
	bc := newBlockchain(db.NewMemDB(), genesisDoc)
	bs, err := bc.Encode()
	require.NoError(t, err)
	bcOut, err := DecodeBlockchain(bs)
	require.True(t, bc.validators.Equal(bcOut.validators))
	require.Equal(t, bc.genesisDoc.GenesisTime, bcOut.genesisDoc.GenesisTime)
	assert.Equal(t, config.JSONString(bc.genesisDoc), config.JSONString(bcOut.genesisDoc))
	require.Equal(t, bc.genesisDoc.Hash(), bcOut.genesisDoc.Hash())
	power := new(big.Int).SetUint64(genesisDoc.Validators[1].Amount)
	id1 := validators[1].PublicKey()
	var flow *big.Int
	for i := 0; i < 100; i++ {
		power := power.Div(power, big.NewInt(2))
		flow, err = bc.AlterPower(id1, power)
		fmt.Println(flow)
		require.NoError(t, err)
		err = bc.CommitBlock(time.Now(), []byte("blockhash"), []byte("apphash"))
		require.NoError(t, err)
		bs, err = bc.Encode()
		require.NoError(t, err)
		bcOut, err = DecodeBlockchain(bs)
		require.True(t, bc.validators.Equal(bcOut.validators))
	}

	// Should have exponentially decayed to 0
	assertZero(t, flow)
	assertZero(t, bc.validators.Prev().Power(id1))
}
