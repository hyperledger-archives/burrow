// Copyright 2019 Monax Industries Limited
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

package storage

import (
	"fmt"

	"github.com/hyperledger/burrow/binary"
	amino "github.com/tendermint/go-amino"
)

var codec = amino.NewCodec()

type CommitID struct {
	Hash    binary.HexBytes
	Version int64
}

func MarshalCommitID(hash []byte, version int64) ([]byte, error) {
	commitID := CommitID{
		Version: version,
		Hash:    hash,
	}
	bs, err := codec.MarshalBinaryBare(commitID)
	if err != nil {
		return nil, fmt.Errorf("MarshalCommitID() could not encode CommitID %v: %v", commitID, err)
	}
	if bs == nil {
		// Normalise zero value to non-nil so we can store it IAVL tree without panic
		return []byte{}, nil
	}
	return bs, nil
}

func UnmarshalCommitID(bs []byte) (*CommitID, error) {
	commitID := new(CommitID)
	err := codec.UnmarshalBinaryBare(bs, commitID)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal CommitID: %v", err)
	}
	return commitID, nil
}

func (cid CommitID) String() string {
	return fmt.Sprintf("Commit{Hash: %v, Version: %v}", cid.Hash, cid.Version)
}
