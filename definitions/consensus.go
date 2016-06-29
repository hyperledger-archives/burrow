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

package definitions

import (
	crypto           "github.com/tendermint/go-crypto"
	p2p              "github.com/tendermint/go-p2p"
	tendermint_types "github.com/tendermint/tendermint/types"
)

// TODO: [ben] explore the value of abstracting the consensus into an interface
// currently we cut a corner here and suffices to have direct calls.

// for now wrap the interface closely around the available Tendermint functions
type ConsensusEngine interface {
	// BlockStore
	Height() int
	LoadBlockMeta(height int) *tendermint_types.BlockMeta

	// Peer-2-Peer
	NodeInfo() *p2p.NodeInfo

	// Private Validator
	PublicValidatorKey() crypto.PubKey
}

// type Communicator interface {
//   // Unicast()
//   Broadcast()
// }
//
// type ConsensusModule interface {
//   Start()
// }
