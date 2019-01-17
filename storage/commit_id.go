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
