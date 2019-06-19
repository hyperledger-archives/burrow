// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package metrics

import (
	"fmt"
	"math"
	"sort"

	"github.com/tendermint/tendermint/types"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/rpc"
	"github.com/prometheus/client_golang/prometheus"
	core_types "github.com/tendermint/tendermint/rpc/core/types"
)

const maxUnconfirmedTxsToFetch = 10000000000
const significantFiguresForSeconds = 3

type HistogramBuilder func(values []float64) (buckets map[float64]uint64, sum float64)

// Exporter is used to store Metrics data and embeds the config struct.
// This is done so that the relevant functions have easy access to the
// user defined runtime configuration when the Collect method is called.
type Exporter struct {
	service                      InfoService
	datum                        *Datum
	chainID                      string
	validatorMoniker             string
	blockSampleSize              uint64
	txPerBlockHistogramBuilder   HistogramBuilder
	timePerBlockHistogramBuilder HistogramBuilder
	logger                       *logging.Logger
}

// Subset of rpc.Service
type InfoService interface {
	Status() (*rpc.ResultStatus, error)
	UnconfirmedTxs(maxTxs int64) (*rpc.ResultUnconfirmedTxs, error)
	Peers() []core_types.Peer
	Blocks(minHeight, maxHeight int64) (*rpc.ResultBlocks, error)
	Stats() acmstate.AccountStatsGetter
}

// Datum is used to store data from all the relevant endpoints
type Datum struct {
	LatestBlockHeight   float64
	UnconfirmedTxs      float64
	TotalPeers          float64
	InboundPeers        float64
	OutboundPeers       float64
	BlockSampleSize     uint64
	TotalTxs            float64
	TxPerBlockBuckets   map[float64]uint64
	TotalTime           float64
	TimePerBlockBuckets map[float64]uint64
	AccountsWithCode    float64
	AccountsWithoutCode float64
}

// Exporter uses the InfoService to provide pre-aggregated metrics of various types that are then passed to prometheus
// as Const metrics rather than being accumulated by individual operations throughout the rest of the Burrow code.
func NewExporter(service InfoService, blockSampleSize int, logger *logging.Logger) (*Exporter, error) {
	chainStatus, err := service.Status()
	if err != nil {
		return nil, fmt.Errorf("NewExporter(): %v", err)
	}
	return &Exporter{
		datum:                        &Datum{},
		service:                      service,
		chainID:                      chainStatus.NodeInfo.Network,
		validatorMoniker:             chainStatus.NodeInfo.Moniker,
		blockSampleSize:              uint64(blockSampleSize),
		txPerBlockHistogramBuilder:   makeHistogramBuilder(identity),
		timePerBlockHistogramBuilder: makeHistogramBuilder(significantFiguresRounder(significantFiguresForSeconds)),
		logger:                       logger.With(structure.ComponentKey, "Metrics_Exporter"),
	}, nil
}

// Describe - loops through the API metrics and passes them to prometheus.Describe
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range MetricDescriptions {
		ch <- m
	}
}

// Collect function, called on by Prometheus Client library
// This function is called when a scrape is performed by requesting /metrics
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	// Scrape the Data
	var err = e.gatherData()
	if err != nil {
		return
	}
	// Set prometheus gauge metrics using the data gathered
	ch <- prometheus.MustNewConstMetric(
		Height,
		prometheus.CounterValue,
		e.datum.LatestBlockHeight,
		e.chainID,
		e.validatorMoniker,
	)
	ch <- prometheus.MustNewConstMetric(
		UnconfirmedTransactions,
		prometheus.GaugeValue,
		e.datum.UnconfirmedTxs,
		e.chainID,
		e.validatorMoniker,
	)
	ch <- prometheus.MustNewConstMetric(
		TotalPeers,
		prometheus.GaugeValue,
		e.datum.TotalPeers,
		e.chainID,
		e.validatorMoniker,
	)
	ch <- prometheus.MustNewConstMetric(
		InboundPeers,
		prometheus.GaugeValue,
		e.datum.InboundPeers,
		e.chainID,
		e.validatorMoniker,
	)
	ch <- prometheus.MustNewConstMetric(
		OutboundPeers,
		prometheus.GaugeValue,
		e.datum.OutboundPeers,
		e.chainID,
		e.validatorMoniker,
	)
	ch <- prometheus.MustNewConstHistogram(
		TxPerBlock,
		e.datum.BlockSampleSize,
		e.datum.TotalTxs,
		e.datum.TxPerBlockBuckets,
		e.chainID,
		e.validatorMoniker,
	)
	ch <- prometheus.MustNewConstHistogram(
		TimePerBlock,
		// Duration between each block in sample
		e.datum.BlockSampleSize-1,
		e.datum.TotalTime,
		e.datum.TimePerBlockBuckets,
		e.chainID,
		e.validatorMoniker,
	)
	ch <- prometheus.MustNewConstMetric(
		Contracts,
		prometheus.GaugeValue,
		e.datum.AccountsWithCode,
		e.chainID,
		e.validatorMoniker,
	)
	ch <- prometheus.MustNewConstMetric(
		Users,
		prometheus.GaugeValue,
		e.datum.AccountsWithoutCode,
		e.chainID,
		e.validatorMoniker,
	)

	e.logger.InfoMsg("All Metrics successfully collected")
}

// gatherData - Collects the data from the API and stores into struct
func (e *Exporter) gatherData() error {
	var err error

	err = e.getStatus()
	if err != nil {
		return err
	}
	err = e.getMemPoolDepth()
	if err != nil {
		return err
	}
	err = e.getPeers()
	if err != nil {
		return err
	}
	blocks, err := e.getBlocks()
	if err != nil {
		return err
	}
	err = e.getTxBuckets(blocks.BlockMetas)
	if err != nil {
		return err
	}
	err = e.getBlockTimeBuckets(blocks.BlockMetas)
	if err != nil {
		return err
	}
	e.getAccountStats()

	return nil
}

// Get status
func (e *Exporter) getStatus() error {
	res, err := e.service.Status()
	if err != nil {
		return err
	}
	e.datum.LatestBlockHeight = float64(res.SyncInfo.LatestBlockHeight)
	return nil
}

// Get unconfirmed_transactions
func (e *Exporter) getMemPoolDepth() error {
	res, err := e.service.UnconfirmedTxs(maxUnconfirmedTxsToFetch)
	if err != nil {
		return err
	}
	e.datum.UnconfirmedTxs = float64(res.NumTxs)
	return nil
}

// Get total peers
func (e *Exporter) getPeers() error {
	peers := e.service.Peers()
	e.datum.TotalPeers = float64(len(peers))
	e.datum.InboundPeers = 0
	e.datum.OutboundPeers = 0

	for _, peer := range peers {
		if peer.IsOutbound {
			e.datum.OutboundPeers += 1
		} else {
			e.datum.InboundPeers += 1
		}
	}

	return nil
}

func (e *Exporter) getBlocks() (*rpc.ResultBlocks, error) {
	var minHeight uint64
	maxHeight := uint64(e.datum.LatestBlockHeight)

	if maxHeight >= e.blockSampleSize {
		minHeight = maxHeight - (e.blockSampleSize - 1)
		e.datum.BlockSampleSize = e.blockSampleSize
	} else {
		minHeight = 1
		e.datum.BlockSampleSize = maxHeight
	}

	res, err := e.service.Blocks(int64(minHeight), int64(maxHeight))
	if err != nil {
		return nil, err
	}

	if len(res.BlockMetas) == 0 {
		return nil, fmt.Errorf("no blocks returned")
	}

	return res, nil
}

// Get transaction buckets
func (e *Exporter) getTxBuckets(blockMetas []*types.BlockMeta) error {
	e.datum.TotalTxs = 0
	e.datum.TxPerBlockBuckets = map[float64]uint64{}
	if len(blockMetas) == 0 {
		return nil
	}
	// Collect number of txs per block as an array of floats
	txsPerBlock := make([]float64, len(blockMetas))
	for i, block := range blockMetas {
		txsPerBlock[i] = float64(block.Header.NumTxs)
	}

	e.datum.TxPerBlockBuckets, e.datum.TotalTxs = e.txPerBlockHistogramBuilder(txsPerBlock)
	return nil
}

func (e *Exporter) getBlockTimeBuckets(blockMetas []*types.BlockMeta) error {
	e.datum.TotalTime = 0
	e.datum.TimePerBlockBuckets = map[float64]uint64{}
	if len(blockMetas) < 2 {
		return nil
	}
	if blockMetas[0].Header.Height == 1 {
		// The LastBlockTime on the first block is the GenesisDoc time! We don't want this
		// in the block duration statistics
		return e.getBlockTimeBuckets(blockMetas[1:])
	}
	blockDurations := make([]float64, len(blockMetas)-1)
	for i := 0; i < len(blockMetas)-1; i++ {
		timeBegan := blockMetas[i].Header.Time
		timeEnded := blockMetas[i+1].Header.Time
		blockDurations[i] = timeEnded.Sub(timeBegan).Seconds()
	}

	e.datum.TimePerBlockBuckets, e.datum.TotalTime = e.timePerBlockHistogramBuilder(blockDurations)
	return nil
}

func (e *Exporter) getAccountStats() {
	stats := e.service.Stats().GetAccountStats()
	e.datum.AccountsWithCode = float64(stats.AccountsWithCode)
	e.datum.AccountsWithoutCode = float64(stats.AccountsWithoutCode)
}

// Returns a function that builds a histogram.
//
// The builder takes a slice of values one for each entity in a sample, sorts it, and computes histogram buckets as
// a map from upper bounds (of values) to cumulative counts (of entities)
// such that count-many entities have value less than or equal to each upper bound.
// Returns this map and the sum of all values.
//
// The smoothing function can be used to round up the upper bounds of buckets so that generated buckets fall on rounder
// more predictable values. This can make querying for a specific bucket easier but not strictly necessary since we can
// use histogram_quantile function in promql to aggregate buckets without needing to know upper bounds ahead of time.
//
// For example if we have the collection (people are entities, numbers of oranges are values)
//
// Fred: 12 oranges
// Annie: 4 oranges
// Paul: 1 orange
//
// We return:
//
// upper bound number of oranges => cumulative number of people (with less than or equal to upper bound number of oranges)
// 1 => 1
// 4 => 2
// 12 => 3
//
// and 17 for the sum
func makeHistogramBuilder(smooth func(float64) float64) HistogramBuilder {
	return func(vals []float64) (buckets map[float64]uint64, sum float64) {
		buckets = make(map[float64]uint64)
		sort.Slice(vals, func(i, j int) bool { return vals[i] < vals[j] })

		var count uint64
		for _, upper := range vals {
			count++
			// Use unsmoothed value for sum so we get true value
			sum += upper
			// Use smoothed upper bound for buckets. Collisions here are fine since the count is cumulative so any
			// earlier upper/count pairs that are overwritten will have their count captured in the overwriting count
			buckets[smooth(upper)] = count
		}
		return buckets, sum
	}
}

func identity(x float64) float64 {
	return x
}

func significantFiguresRounder(n int) func(float64) float64 {
	// exponent base 10 for this many sig figs
	sfPow := float64(n - 1)
	return func(x float64) float64 {
		// Floor of exponent to take us to 1 sig fig
		log10x := math.Floor(math.Log10(x))
		// Power of 10 to scale us to n sig figs
		fac := math.Pow(10, sfPow-log10x)
		// Scale to n sig figs, drop digits to right of decimal point, then scale back
		return math.Round(x*fac) / fac
	}
}
