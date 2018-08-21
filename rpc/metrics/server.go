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
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/lib/server"
)

// Exporter is used to store Metrics data and embeds the config struct.
// This is done so that the relevant functions have easy access to the
// user defined runtime configuration when the Collect method is called.
type Exporter struct {
	burrowMetrics    map[string]*prometheus.Desc
	service          *rpc.Service
	logger           *logging.Logger
	datum            *Datum
	chainID          string
	validatorMoniker string
	blockSampleSize  uint64
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
	TxPerBlockBuckets   map[float64]float64
	TotalTime           float64
	TimePerBlockBuckets map[float64]float64
}

func StartServer(service *rpc.Service, pattern, listenAddress string, blockSampleSize uint64,
	logger *logging.Logger) (*http.Server, error) {

	// instantiate metrics and variables we do not expect to change during runtime
	chainStatus, _ := service.Status()
	exporter := Exporter{
		burrowMetrics:    AddMetrics(),
		datum:            &Datum{},
		logger:           logger.With(structure.ComponentKey, "Metrics_Exporter"),
		service:          service,
		chainID:          chainStatus.NodeInfo.Network,
		validatorMoniker: chainStatus.NodeInfo.Moniker,
		blockSampleSize:  blockSampleSize,
	}

	// Register Metrics from each of the endpoints
	// This invokes the Collect method through the prometheus client libraries.
	prometheus.MustRegister(&exporter)

	mux := http.NewServeMux()
	mux.Handle(pattern, server.RecoverAndLogHandler(prometheus.Handler(), logger))

	srv, err := server.StartHTTPServer(listenAddress, mux, logger)
	if err != nil {
		return nil, err
	}
	return srv, nil
}

// AddMetrics - Add's all of the metrics to a map of strings, returns the map.
func AddMetrics() map[string]*prometheus.Desc {
	burrowMetrics := make(map[string]*prometheus.Desc)

	burrowMetrics["Height"] = prometheus.NewDesc(
		prometheus.BuildFQName("burrow", "chain", "block_height"),
		"Current block height",
		[]string{"chain_id", "moniker"}, nil,
	)
	burrowMetrics["Time Per Block"] = prometheus.NewDesc(
		prometheus.BuildFQName("burrow", "chain", "block_time"),
		"Summary metric of nanoseconds per block",
		[]string{"chain_id", "moniker"}, nil,
	)
	burrowMetrics["Unconfirmed Transactions"] = prometheus.NewDesc(
		prometheus.BuildFQName("burrow", "transactions", "in_mempool"),
		"Current depth of the mempool",
		[]string{"chain_id", "moniker"}, nil,
	)
	burrowMetrics["Tx Per Block"] = prometheus.NewDesc(
		prometheus.BuildFQName("burrow", "transactions", "per_block"),
		"Summary metric of transactions per block",
		[]string{"chain_id", "moniker"}, nil,
	)
	burrowMetrics["Total Peers"] = prometheus.NewDesc(
		prometheus.BuildFQName("burrow", "peers", "total"),
		"Current total peers",
		[]string{"chain_id", "moniker"}, nil,
	)
	burrowMetrics["Inbound Peers"] = prometheus.NewDesc(
		prometheus.BuildFQName("burrow", "peers", "inbound"),
		"Current inbound peers",
		[]string{"chain_id", "moniker"}, nil,
	)
	burrowMetrics["Outbound Peers"] = prometheus.NewDesc(
		prometheus.BuildFQName("burrow", "peers", "outbound"),
		"Current outbound peers",
		[]string{"chain_id", "moniker"}, nil,
	)

	return burrowMetrics
}

// Describe - loops through the API metrics and passes them to prometheus.Describe
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.burrowMetrics {
		ch <- m
	}
}

// Collect function, called on by Prometheus Client library
// This function is called when a scrape is peformed on the /metrics page
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	// Scrape the Data
	var err = e.gatherData()
	if err != nil {
		return
	}

	// Set prometheus gauge metrics using the data gathered
	ch <- prometheus.MustNewConstMetric(
		e.burrowMetrics["Height"],
		prometheus.CounterValue,
		e.datum.LatestBlockHeight,
		e.chainID,
		e.validatorMoniker,
	)
	ch <- prometheus.MustNewConstMetric(
		e.burrowMetrics["Unconfirmed Transactions"],
		prometheus.GaugeValue,
		e.datum.UnconfirmedTxs,
		e.chainID,
		e.validatorMoniker,
	)
	ch <- prometheus.MustNewConstMetric(
		e.burrowMetrics["Total Peers"],
		prometheus.GaugeValue,
		e.datum.TotalPeers,
		e.chainID,
		e.validatorMoniker,
	)
	ch <- prometheus.MustNewConstMetric(
		e.burrowMetrics["Inbound Peers"],
		prometheus.GaugeValue,
		e.datum.InboundPeers,
		e.chainID,
		e.validatorMoniker,
	)
	ch <- prometheus.MustNewConstMetric(
		e.burrowMetrics["Outbound Peers"],
		prometheus.GaugeValue,
		e.datum.OutboundPeers,
		e.chainID,
		e.validatorMoniker,
	)
	ch <- prometheus.MustNewConstSummary(
		e.burrowMetrics["Tx Per Block"],
		e.datum.BlockSampleSize,
		e.datum.TotalTxs,
		e.datum.TxPerBlockBuckets,
		e.chainID,
		e.validatorMoniker,
	)
	ch <- prometheus.MustNewConstSummary(
		e.burrowMetrics["Time Per Block"],
		e.datum.BlockSampleSize,
		e.datum.TotalTime,
		e.datum.TimePerBlockBuckets,
		e.chainID,
		e.validatorMoniker,
	)

	e.logger.InfoMsg("All Metrics successfully collected")
}
