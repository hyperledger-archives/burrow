package metrics

import "github.com/prometheus/client_golang/prometheus"

var MetricDescriptions = make(map[string]*prometheus.Desc)

var (
	Height = newDesc(
		prometheus.BuildFQName("burrow", "chain", "block_height"),
		"Current block height",
		[]string{"chain_id", "moniker"})

	TimePerBlock = newDesc(
		prometheus.BuildFQName("burrow", "chain", "block_time"),
		"Histogram metric of block duration",
		[]string{"chain_id", "moniker"})

	UnconfirmedTransactions = newDesc(
		prometheus.BuildFQName("burrow", "transactions", "in_mempool"),
		"Current depth of the mempool",
		[]string{"chain_id", "moniker"})

	TxPerBlock = newDesc(
		prometheus.BuildFQName("burrow", "transactions", "per_block"),
		"Histogram metric of transactions per block",
		[]string{"chain_id", "moniker"})

	TotalPeers = newDesc(
		prometheus.BuildFQName("burrow", "peers", "total"),
		"Current total peers",
		[]string{"chain_id", "moniker"})

	InboundPeers = newDesc(
		prometheus.BuildFQName("burrow", "peers", "inbound"),
		"Current inbound peers",
		[]string{"chain_id", "moniker"})

	OutboundPeers = newDesc(
		prometheus.BuildFQName("burrow", "peers", "outbound"),
		"Current outbound peers",
		[]string{"chain_id", "moniker"})

	Contracts = newDesc(
		prometheus.BuildFQName("burrow", "accounts", "contracts"),
		"Current contracts on the chain",
		[]string{"chain_id", "moniker"})

	Users = newDesc(
		prometheus.BuildFQName("burrow", "accounts", "users"),
		"Current users on the chain",
		[]string{"chain_id", "moniker"})
)

func newDesc(fqName, help string, variableLabels []string) *prometheus.Desc {
	desc := prometheus.NewDesc(fqName, help, variableLabels, nil)
	MetricDescriptions[fqName] = desc
	return desc
}
