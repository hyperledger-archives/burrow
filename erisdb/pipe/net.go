package pipe

import (
	"github.com/eris-ltd/eris-db/tendermint/tendermint/p2p"
)

// The net struct.
type network struct {
	p2pSwitch *p2p.Switch
}

func newNetwork(p2pSwitch *p2p.Switch) *network {
	return &network{p2pSwitch}
}

//-----------------------------------------------------------------------------

// Get the complete net info.
func (this *network) Info() (*NetworkInfo, error) {
	version := config.GetString("version")
	moniker := config.GetString("moniker")
	listening := this.p2pSwitch.IsListening()
	listeners := []string{}
	for _, listener := range this.p2pSwitch.Listeners() {
		listeners = append(listeners, listener.String())
	}
	peers := make([]*Peer, 0)
	for _, peer := range this.p2pSwitch.Peers().List() {
		p := &Peer{peer.NodeInfo, peer.IsOutbound()}
		peers = append(peers, p)
	}
	return &NetworkInfo{
		version,
		moniker,
		listening,
		listeners,
		peers,
	}, nil
}

// Get the client version
func (this *network) ClientVersion() (string, error) {
	return config.GetString("version"), nil
}

// Get the moniker
func (this *network) Moniker() (string, error) {
	return config.GetString("moniker"), nil
}

// Is the network currently listening for connections.
func (this *network) Listening() (bool, error) {
	return this.p2pSwitch.IsListening(), nil
}

// Is the network currently listening for connections.
func (this *network) Listeners() ([]string, error) {
	listeners := []string{}
	for _, listener := range this.p2pSwitch.Listeners() {
		listeners = append(listeners, listener.String())
	}
	return listeners, nil
}

// Get a list of all peers.
func (this *network) Peers() ([]*Peer, error) {
	peers := make([]*Peer, 0)
	for _, peer := range this.p2pSwitch.Peers().List() {
		p := &Peer{peer.NodeInfo, peer.IsOutbound()}
		peers = append(peers, p)
	}
	return peers, nil
}

// Get a peer. TODO Need to do something about the address.
func (this *network) Peer(address string) (*Peer, error) {
	peer := this.p2pSwitch.Peers().Get(address)
	return &Peer{peer.NodeInfo, peer.IsOutbound()}, nil
}
