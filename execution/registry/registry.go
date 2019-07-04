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

func (rn *RegisteredNode) String() string {
	return fmt.Sprintf("RegisterNode{%v -> %v @ %v}", rn.PublicKey, rn.ID, rn.NetAddress)
}

type Reader interface {
	GetNetworkRegistry() (map[crypto.Address]*RegisteredNode, error)
}

type Writer interface {
	RegisterNode(val crypto.Address, regNode *RegisteredNode) error
}

type ReaderWriter interface {
	Reader
	Writer
}

type Iterable interface {
	IterateNodes(consumer func(crypto.Address, *RegisteredNode) error) (err error)
}

type IterableReader interface {
	Iterable
	Reader
}

type IterableReaderWriter interface {
	Iterable
	ReaderWriter
}
