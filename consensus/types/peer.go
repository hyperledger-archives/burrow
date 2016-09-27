package types

import "github.com/tendermint/go-p2p"

type Peer struct {
	NodeInfo *p2p.NodeInfo `json:"node_info"`
	IsOutbound   bool `json:"is_outbound"`
}
