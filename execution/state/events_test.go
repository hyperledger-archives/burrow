package state

import (
	bin "encoding/binary"
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/crypto"

	"github.com/hyperledger/burrow/execution/exec"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
)

func TestWriteState_AddBlock(t *testing.T) {
	s := NewState(dbm.NewMemDB())
	height := uint64(100)
	numTxs := uint64(5)
	events := uint64(10)
	block := mkBlock(height, numTxs, events)
	_, _, err := s.Update(func(ws Updatable) error {
		return ws.AddBlock(block)
	})
	require.NoError(t, err)

	txIndex := uint64(0)
	eventIndex := uint64(0)
	err = s.IterateStreamEvents(&height, &height,
		func(ev *exec.StreamEvent) error {
			switch {
			case ev.BeginTx != nil:
				eventIndex = 0
			case ev.Event != nil:
				require.Equal(t, mkEvent(height, txIndex, eventIndex).Header.TxHash.String(),
					ev.Event.Header.TxHash.String(), "event TxHash mismatch at tx #%d event #%d",
					txIndex, eventIndex)
				eventIndex++
			case ev.EndTx != nil:
				txIndex++
			}
			return nil
		})
	require.NoError(t, err)
	require.Equal(t, numTxs, txIndex, "should have observed all txs")
	// non-increasing events
	_, _, err = s.Update(func(ws Updatable) error {
		return nil
	})
	require.NoError(t, err)

	txExecutions, err := s.TxsAtHeight(height)
	require.NoError(t, err)
	require.NotNil(t, txExecutions)
	require.Equal(t, numTxs, uint64(len(txExecutions)))
}

func TestNestedTxs(t *testing.T) {
	s := NewState(dbm.NewMemDB())
	height := uint64(2)
	numTxs := uint64(4)
	events := uint64(2)
	nesting := uint64(3)
	block := mkBlock(height, numTxs, events)
	txes := block.TxExecutions
	// Deeply nest transactions inside block
	for i := uint64(0); i < nesting; i++ {
		var next []*exec.TxExecution
		for _, txe := range txes {
			next = append(next, nestTxs(txe, height, events+i, numTxs+i)...)
		}
		txes = next
	}
	_, _, err := s.Update(func(ws Updatable) error {
		return ws.AddBlock(block)
	})
	require.NoError(t, err)
	txes, err = s.TxsAtHeight(height)
	require.NoError(t, err)
	txsCount := deepCountTxs(block.TxExecutions)
	require.Equal(t, txsCount, deepCountTxs(txes))
	// There is a geometric-arithmetic sum here... but empiricism FTW
	require.Equal(t, 580, txsCount)
}

func TestReadState_TxByHash(t *testing.T) {
	s := NewState(dbm.NewMemDB())
	maxHeight := uint64(3)
	numTxs := uint64(4)
	events := uint64(2)
	for height := uint64(0); height < maxHeight; height++ {
		block := mkBlock(height, numTxs, events)
		_, _, err := s.Update(func(ws Updatable) error {
			return ws.AddBlock(block)
		})
		require.NoError(t, err)
	}

	hashSet := make(map[string]bool)
	for height := uint64(0); height < maxHeight; height++ {
		for txIndex := uint64(0); txIndex < numTxs; txIndex++ {
			// Find this tx
			tx := mkTx(height, txIndex, events)
			txHash := tx.TxHash.String()
			// Check we have no duplicates (indicates problem with how we are generating hashes for these tests
			require.False(t, hashSet[txHash], "should be no duplicate tx hashes")
			hashSet[txHash] = true
			// Try and pull the Tx by its hash
			txOut, err := s.TxByHash(tx.TxHash)
			require.NoError(t, err)
			require.NotNil(t, txOut, "should retrieve non-nil transaction by TxHash %v", tx.TxHash)
			// Make sure we get the same tx
			require.Equal(t, txHash, txOut.TxHash.String(), "TxHash does not match as string")
			require.Equal(t, source.JSONString(tx), source.JSONString(txOut))
		}
	}
}

func deepCountTxs(txes []*exec.TxExecution) int {
	sum := len(txes)
	for _, txe := range txes {
		sum += deepCountTxs(txe.TxExecutions)
	}
	return sum
}

func nestTxs(txe *exec.TxExecution, height, events, numTxs uint64) []*exec.TxExecution {
	txes := make([]*exec.TxExecution, numTxs)
	for i := uint64(0); i < numTxs; i++ {
		txes[i] = mkTx(height, i, events)
		txe.TxExecutions = append(txe.TxExecutions, txes[i])
	}
	return txes
}

func mkBlock(height, numTxs, events uint64) *exec.BlockExecution {
	be := &exec.BlockExecution{
		Height: height,
	}
	for ti := uint64(0); ti < numTxs; ti++ {
		txe := mkTx(height, ti, events)
		be.TxExecutions = append(be.TxExecutions, txe)
	}
	return be
}

func mkTx(height, txIndex, events uint64) *exec.TxExecution {
	hash := make([]byte, 32)
	bin.BigEndian.PutUint64(hash[:8], height)
	bin.BigEndian.PutUint64(hash[8:16], txIndex)
	bin.BigEndian.PutUint64(hash[16:24], events)
	txe := &exec.TxExecution{
		TxHeader: &exec.TxHeader{
			TxHash: hash,
			Height: height,
			Index:  txIndex,
		},
	}
	for e := uint64(0); e < events; e++ {
		txe.Events = append(txe.Events, mkEvent(height, txIndex, e))
	}
	return txe
}

func mkEvent(height, tx, index uint64) *exec.Event {
	return &exec.Event{
		Header: &exec.Header{
			Height:  height,
			Index:   index,
			TxHash:  crypto.Keccak256([]byte(fmt.Sprintf("txhash%v%v%v", height, tx, index))),
			EventID: fmt.Sprintf("eventID: %v%v%v", height, tx, index),
		},
		Log: &exec.LogEvent{
			Address: crypto.Address{byte(height), byte(index)},
			Topics:  []binary.Word256{{1, 2, 3}},
		},
	}
}

func BenchmarkAddBlockAndIterator(b *testing.B) {
	s := NewState(dbm.NewMemDB())
	numTxs := uint64(5)
	events := uint64(10)
	for height := uint64(0); height < 2000; height++ {
		block := mkBlock(height, numTxs, events)
		_, _, err := s.Update(func(ws Updatable) error {
			return ws.AddBlock(block)
		})
		require.NoError(b, err)
	}
	err := s.IterateStreamEvents(nil, nil,
		func(ev *exec.StreamEvent) error {
			return nil
		})
	require.NoError(b, err)
}
