package metrics

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	io_prometheus_client "github.com/prometheus/client_model/go"

	"github.com/hyperledger/burrow/bcm"

	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/types"
)

func TestExporter_Collect_Histogram(t *testing.T) {
	ch := make(chan prometheus.Metric)
	done := make(chan error)
	metrics := make(map[string]*io_prometheus_client.Metric)

	// Pull up a mock service
	is := infoService()
	sampleSize := 100
	exporter, err := NewExporter(is, sampleSize, logging.NewNoopLogger())
	require.NoError(t, err)

	// Start waiting for us to push metrics to channel from Collect()
	go func() {
		defer close(done)
		for m := range ch {
			model := new(io_prometheus_client.Metric)
			err := m.Write(model)
			if err != nil {
				done <- err
				return
			}
			metrics[m.Desc().String()] = model
		}
	}()

	// Push the metrics and wait until we have written them into their models
	exporter.Collect(ch)
	close(ch)
	require.NoError(t, <-done)

	// Extract the relevant histogram values as floats
	numTxsPerBlock := make([]float64, sampleSize)
	timePerBlock := make([]float64, sampleSize-1)

	var blockTime time.Time
	for i, block := range is.BlockMetas[len(is.BlockMetas)-sampleSize:] {
		if i > 0 {
			timePerBlock[i-1] = block.Header.Time.Sub(blockTime).Seconds()
		}
		blockTime = block.Header.Time
		numTxsPerBlock[i] = float64(block.Header.NumTxs)
	}

	// Pull out the histograms and make sure they check out
	txPerBlockMetric := metrics[TxPerBlock.String()]
	require.NotNil(t, txPerBlockMetric)
	require.NotNil(t, txPerBlockMetric.Histogram)
	verifyHistogram(t, txPerBlockMetric.Histogram, numTxsPerBlock, identity)

	timePerBlockMetric := metrics[TimePerBlock.String()]
	require.NotNil(t, timePerBlockMetric)
	require.NotNil(t, timePerBlockMetric.Histogram)
	verifyHistogram(t, timePerBlockMetric.Histogram, timePerBlock, significantFiguresRounder(significantFiguresForSeconds))
}

func TestSignificantFigures(t *testing.T) {
	f := significantFiguresRounder(3)
	assert.Equal(t, float64(21400), f(21432))
	assert.Equal(t, float64(2140), f(2143))
	assert.Equal(t, float64(0.1), f(0.1))
	assert.Equal(t, float64(0.123), f(0.123))
	assert.Equal(t, float64(0.123), f(0.1234))
	assert.Equal(t, float64(0.124), f(0.1235))
	assert.Equal(t, float64(10), f(10))
	assert.Equal(t, float64(100), f(100))
	assert.Equal(t, float64(0.000344), f(0.00034363))
	assert.Equal(t, float64(0.00123), f(0.001234))

	f = significantFiguresRounder(1)
	assert.Equal(t, float64(20000), f(21432))
	assert.Equal(t, float64(0.00009), f(0.0000911123))
}

func verifyHistogram(t *testing.T, histogram *io_prometheus_client.Histogram, values []float64,
	smooth func(float64) float64) {
	buckets := histogram.Bucket
	// Get cumulative totals for each bucket
	counts := make([]int64, len(buckets))
	// Get total sum of all values
	var sum float64
	for _, value := range values {
		sum += value
		// Use the upper bounds to bin the value
		for i, bucket := range buckets {
			if smooth(value) <= *bucket.UpperBound {
				counts[i] += 1
			}
		}
	}
	// The cumulative counts we have just calculated for each bucket should match
	for i, bucket := range buckets {
		assert.Equal(t, counts[i], int64(*bucket.CumulativeCount))
	}
	assert.Equal(t, counts[len(buckets)-1], int64(*histogram.SampleCount))
	assert.InDelta(t, sum, *histogram.SampleSum, 0.01)
}

func infoService() *constInfo {
	numBlocks := int64(1000)
	return &constInfo{
		ResultStatus: &rpc.ResultStatus{
			NodeInfo: &tendermint.NodeInfo{
				Network: "TestChain",
				Moniker: "TestNode",
			},
			SyncInfo: &bcm.SyncInfo{
				LatestBlockHeight: uint64(numBlocks),
			},
		},
		ResultUnconfirmedTxs: &rpc.ResultUnconfirmedTxs{},
		BlockMetas:           genBlocks(numBlocks),
	}
}

func genBlocks(n int64) []*types.BlockMeta {
	bms := make([]*types.BlockMeta, n)
	rnd := rand.New(rand.NewSource(n))
	blockTime := time.Unix(rnd.Int63()>>32, 0)
	for i := int64(0); i < n; i++ {
		blockDuration := time.Second*2 + time.Millisecond*time.Duration(rnd.Int63()>>48)
		blockTime = blockTime.Add(blockDuration)
		bms[i] = blockMeta(i+1, rnd.Int63()>>56, blockTime)
	}
	return bms
}

func blockMeta(height, numTxs int64, blockTime time.Time) *types.BlockMeta {
	return &types.BlockMeta{
		Header: types.Header{
			Height: height,
			NumTxs: numTxs,
			Time:   blockTime,
		},
	}
}
