package pipe

import ()

// TODO-RPC!

// The net struct.
type network struct {
}

func newNetwork() *network {
	return &network{}
}

//-----------------------------------------------------------------------------

// Get the complete net info.
func (this *network) Info() (*NetworkInfo, error) {
	return &NetworkInfo{}, nil
}

// Get the client version
func (this *network) ClientVersion() (string, error) {
	return "not-fully-loaded-yet", nil
}

// Get the moniker
func (this *network) Moniker() (string, error) {
	return "rekinom", nil
}

// Is the network currently listening for connections.
func (this *network) Listening() (bool, error) {
	return false, nil
}

// Is the network currently listening for connections.
func (this *network) Listeners() ([]string, error) {
	return []string{}, nil
}

// Get a list of all peers.
func (this *network) Peers() ([]*Peer, error) {
	return []*Peer{}, nil
}

// Get a peer. TODO Need to do something about the address.
func (this *network) Peer(address string) (*Peer, error) {
	return &Peer{}, nil
}
