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
package shared

import (
	consensus_types "github.com/eris-ltd/eris-db/consensus/types"
	"github.com/eris-ltd/eris-db/definitions"
)

//-----------------------------------------------------------------------------

// Get the complete pipe info.
func NetInfo(pipe definitions.Pipe) *NetworkInfo {
	thisNode := pipe.GetConsensusEngine().NodeInfo()
	listeners := []string{}
	for _, listener := range pipe.GetConsensusEngine().Listeners() {
		listeners = append(listeners, listener.String())
	}
	return &NetworkInfo{
		ClientVersion: thisNode.Version,
		Moniker:       thisNode.Moniker,
		Listening:     pipe.GetConsensusEngine().IsListening(),
		Listeners:     listeners,
		Peers:         pipe.GetConsensusEngine().Peers(),
	}
}

// Get the client version
func ClientVersion(pipe definitions.Pipe) string {
	return pipe.GetConsensusEngine().NodeInfo().Version
}

// Get the moniker
func Moniker(pipe definitions.Pipe) string {
	return pipe.GetConsensusEngine().NodeInfo().Moniker
}

// Is the network currently listening for connections.
func Listening(pipe definitions.Pipe) bool {
	return pipe.GetConsensusEngine().IsListening()
}

// Is the network currently listening for connections.
func Listeners(pipe definitions.Pipe) []string {
	listeners := []string{}
	for _, listener := range pipe.GetConsensusEngine().Listeners() {
		listeners = append(listeners, listener.String())
	}
	return listeners
}

func Peer(pipe definitions.Pipe, address string) *consensus_types.Peer {
	for _, peer := range pipe.GetConsensusEngine().Peers() {
		if peer.NodeInfo.RemoteAddr == address {
			return peer
		}
	}
	return nil
}

// NetworkInfo
type NetworkInfo struct {
	ClientVersion string                  `json:"client_version"`
	Moniker       string                  `json:"moniker"`
	Listening     bool                    `json:"listening"`
	Listeners     []string                `json:"listeners"`
	Peers         []*consensus_types.Peer `json:"peers"`
}
