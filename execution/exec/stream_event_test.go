package exec

import (
	"testing"
	"time"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/abci/types"
)

var genesisDoc, accounts, _ = genesis.NewDeterministicGenesis(345234523).GenesisDoc(10, 0)

func TestTxExecution(t *testing.T) {
	txe := NewTxExecution(txs.Enclose(genesisDoc.ChainID(), newCallTx(0, 1)))

	stack := new(TxStack)
	var txeOut *TxExecution
	var err error

	for _, ev := range txe.StreamEvents() {
		txeOut, err = stack.Consume(ev)
		require.NoError(t, err)
		if txeOut != nil {
			require.Equal(t, txe, txeOut)
		}
	}

	assert.NotNil(t, txeOut, "should have consumed input TxExecution")
}

func TestConsumeBlockExecution(t *testing.T) {
	height := int64(234242)
	be := &BlockExecution{
		Header: &types.Header{
			ChainID: genesisDoc.ChainID(),
			AppHash: crypto.Keccak256([]byte("hashily")),
			Time:    time.Now(),
			Height:  height,
		},
		Height: uint64(height),
		TxExecutions: []*TxExecution{
			NewTxExecution(txs.Enclose(genesisDoc.ChainID(), newCallTx(0, 3))),
			NewTxExecution(txs.Enclose(genesisDoc.ChainID(), newCallTx(0, 2))),
			NewTxExecution(txs.Enclose(genesisDoc.ChainID(), newCallTx(2, 1))),
		},
	}

	stack := new(BlockAccumulator)
	var beOut *BlockExecution
	var err error
	for _, ev := range be.StreamEvents() {
		beOut, err = stack.Consume(ev)
		require.NoError(t, err)
		if beOut != nil {
			require.Equal(t, be, beOut)
		}
	}
	assert.NotNil(t, beOut, "should have consumed input BlockExecution")
}

func newCallTx(fromIndex, toIndex int) *payload.CallTx {
	from := accounts[fromIndex]
	to := accounts[toIndex].GetAddress()
	return payload.NewCallTxWithSequence(from.GetPublicKey(), &to, []byte{1, 2, 3}, 324, 34534534, 23, 1)
}
