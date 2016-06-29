// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

// Net is part of the pipe for ErisMint and provides the implementation
// for the pipe to call into the ErisMint application
package erismint

import (
  core_types           "github.com/eris-ltd/eris-db/core/types"
)

// TODO-RPC!

// The network structure
type network struct {
}

func newNetwork() *network {
	return &network{}
}

//------------------------------------------------------------------------------
// Tendermint Pipe implementation



//-----------------------------------------------------------------------------
// Eris-DB v0 Pipe implementation

// Get the complete net info.
func (this *network) Info() (*core_types.NetworkInfo, error) {
	return &core_types.NetworkInfo{}, nil
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
func (this *network) Peers() ([]*core_types.Peer, error) {
	return []*core_types.Peer{}, nil
}

// Get a peer. TODO Need to do something about the address.
func (this *network) Peer(address string) (*core_types.Peer, error) {
	return &core_types.Peer{}, nil
}

//------------------------------------------------------------------------------
//
