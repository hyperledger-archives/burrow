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
package pipes

import (
	core_types "github.com/eris-ltd/eris-db/core/types"
	"github.com/eris-ltd/eris-db/definitions"
)

//-----------------------------------------------------------------------------

// Get the complete pipe info.
func NetInfo(pipe definitions.Pipe) (*core_types.NetworkInfo, error) {
	thisNode := pipe.GetConsensusEngine().NodeInfo()
	listeners := []string{}
	for _, listener := range pipe.GetConsensusEngine().Listeners() {
		listeners = append(listeners, listener.String())
	}
	peers := make([]*core_types.Peer, 0)
	for _, peer := range pipe.GetConsensusEngine().Peers() {
		p := &core_types.Peer{
			NodeInfo:   &peer.NodeInfo,
			IsOutbound: peer.IsOutbound,
		}
		peers = append(peers, p)
	}
	return &core_types.NetworkInfo{
		ClientVersion: thisNode.Version,
		Moniker:       thisNode.Moniker,
		Listening:     pipe.GetConsensusEngine().IsListening(),
		Listeners:     listeners,
		Peers:         peers,
	}, nil
}

// Get the client version
func ClientVersion(pipe definitions.Pipe) (string, error) {
	return pipe.GetConsensusEngine().NodeInfo().Version, nil
}

// Get the moniker
func Moniker(pipe definitions.Pipe) (string, error) {
	return pipe.GetConsensusEngine().NodeInfo().Moniker, nil
}

// Is the network currently listening for connections.
func Listening(pipe definitions.Pipe) (bool, error) {
	return pipe.GetConsensusEngine().IsListening(), nil
}

// Is the network currently listening for connections.
func Listeners(pipe definitions.Pipe) ([]string, error) {
	listeners := []string{}
	for _, listener := range pipe.GetConsensusEngine().Listeners() {
		listeners = append(listeners, listener.String())
	}
	return listeners, nil
}

// Get a list of all peers.
func Peers(pipe definitions.Pipe) ([]*core_types.Peer, error) {
	peers := make([]*core_types.Peer, 0)
	for _, peer := range pipe.GetConsensusEngine().Peers() {
		p := &core_types.Peer{
			NodeInfo:   &peer.NodeInfo,
			IsOutbound: peer.IsOutbound,
		}
		peers = append(peers, p)
	}
	return peers, nil
}

func  Peer(pipe definitions.Pipe, address string) (*core_types.Peer, error) {
	for _, peer := range pipe.GetConsensusEngine().Peers() {
		if peer.RemoteAddr == address {
			return &core_types.Peer{
				NodeInfo: &peer.NodeInfo,
				IsOutbound: peer.IsOutbound,
			}, nil
		}
	}
	return nil, nil
}