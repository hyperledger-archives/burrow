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

package shared

import (
	consensus_types "github.com/monax/burrow/consensus/types"
	"github.com/monax/burrow/definitions"
)

// Net is part of the pipe for BurrowMint and provides the implementation
// for the pipe to call into the BurrowMint application

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
