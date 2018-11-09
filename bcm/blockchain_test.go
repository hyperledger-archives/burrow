package bcm

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging/logconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/db"
)

var big0 = big.NewInt(0)

func TestBlockchain_Encode(t *testing.T) {
	genesisDoc, _, validators := genesis.NewDeterministicGenesis(234).
		GenesisDoc(5, true, 232, 3, true, 34)
	bc := newBlockchain(db.NewMemDB(), genesisDoc)
	bs, err := bc.Encode()
	require.NoError(t, err)
	bcOut, err := DecodeBlockchain(bs)
	require.True(t, bc.validatorCache.Equal(bcOut.validatorCache))
	require.Equal(t, bc.genesisDoc.GenesisTime, bcOut.genesisDoc.GenesisTime)
	assert.Equal(t, logconfig.JSONString(bc.genesisDoc), logconfig.JSONString(bcOut.genesisDoc))
	require.Equal(t, bc.genesisDoc.Hash(), bcOut.genesisDoc.Hash())
	power := new(big.Int).SetUint64(genesisDoc.Validators[1].Amount)
	id1 := validators[1].GetPublicKey()
	var flow *big.Int
	for i := 0; i < 100; i++ {
		power := power.Div(power, big.NewInt(2))
		flow, err = bc.ValidatorWriter().AlterPower(id1, power)
		fmt.Println(flow)
		require.NoError(t, err)
		_, _, err = bc.CommitBlock(time.Now(), []byte("blockhash"), []byte("apphash"))
		require.NoError(t, err)
		bs, err = bc.Encode()
		require.NoError(t, err)
		bcOut, err = DecodeBlockchain(bs)
		require.True(t, bc.validatorCache.Equal(bcOut.validatorCache))
	}

	// Should have exponentially decayed to 0
	assertZero(t, flow)
	assertZero(t, bc.validatorCache.Power(id1.GetAddress()))
}

// Since we have -0 and 0 with big.Int due to its representation with a neg flag
func assertZero(t testing.TB, i *big.Int) {
	assert.True(t, big0.Cmp(i) == 0, "expected 0 but got %v", i)
}
