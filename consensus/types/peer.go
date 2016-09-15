package types

import "github.com/tendermint/go-p2p"

type Peer struct {
	p2p.NodeInfo `json:"node_info"`
	IsOutbound   bool `json:"is_outbound"`
}
