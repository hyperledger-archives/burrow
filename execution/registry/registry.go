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

package registry

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
)

func AuthorizedPeersProvider(state IterableReader) func() ([]string, []string) {
	return func() ([]string, []string) {
		var peerIDs, peerAddrs []string

		for _, node := range state.GetNodes() {
			peerIDs = append(peerIDs, node.TendermintNodeID)
			peerAddrs = append(peerAddrs, node.NetworkAddress)
		}

		return peerIDs, peerAddrs
	}
}

func (rn *NodeIdentity) String() string {
	return fmt.Sprintf("RegisterNode{%v -> %v @ %v}", rn.ValidatorPublicKey, rn.TendermintNodeID, rn.NetworkAddress)
}

type NodeList map[crypto.Address]*NodeIdentity

type Reader interface {
	GetNode(crypto.Address) (*NodeIdentity, error)
	GetNodes() NodeList
}

type Writer interface {
	// Updates the node, creating it if it does not exist
	UpdateNode(crypto.Address, *NodeIdentity) error
	// Remove the proposal by hash
	RemoveNode(crypto.Address) error
}

type ReaderWriter interface {
	Reader
	Writer
}

type Iterable interface {
	IterateNodes(consumer func(crypto.Address, *NodeIdentity) error) (err error)
}

type IterableReader interface {
	Iterable
	Reader
}

type IterableReaderWriter interface {
	Iterable
	ReaderWriter
}
