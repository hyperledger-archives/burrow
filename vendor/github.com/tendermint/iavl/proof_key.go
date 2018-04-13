package iavl

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
)

// KeyProof represents a proof of existence or absence of a single key.
type KeyProof interface {
	// Verify verfies the proof is valid. To verify absence,
	// the value should be nil.
	Verify(key, value, root []byte) error

	// Root returns the root hash of the proof.
	Root() []byte

	// Serialize itself
	Bytes() []byte
}

const (
	// Used for serialization of proofs.
	keyExistsMagicNumber = 0x50
	keyAbsentMagicNumber = 0x51
)

// KeyExistsProof represents a proof of existence of a single key.
type KeyExistsProof struct {
	RootHash cmn.HexBytes `json:"root_hash"`
	Version  int64        `json:"version"`

	*PathToKey `json:"path"`
}

func (proof *KeyExistsProof) Root() []byte {
	return proof.RootHash
}

// Verify verifies the proof is valid and returns an error if it isn't.
func (proof *KeyExistsProof) Verify(key []byte, value []byte, root []byte) error {
	if !bytes.Equal(proof.RootHash, root) {
		return errors.WithStack(ErrInvalidRoot)
	}
	if key == nil || value == nil {
		return errors.WithStack(ErrInvalidInputs)
	}
	return proof.PathToKey.verify(proofLeafNode{key, value, proof.Version}.Hash(), root)
}

// Bytes returns a go-wire binary serialization
func (proof *KeyExistsProof) Bytes() []byte {
	return append([]byte{keyExistsMagicNumber}, wire.BinaryBytes(proof)...)
}

// readKeyExistsProof will deserialize a KeyExistsProof from bytes.
func readKeyExistsProof(data []byte) (*KeyExistsProof, error) {
	proof := new(KeyExistsProof)
	err := wire.ReadBinaryBytes(data, &proof)
	return proof, err
}

///////////////////////////////////////////////////////////////////////////////

// KeyAbsentProof represents a proof of the absence of a single key.
type KeyAbsentProof struct {
	RootHash cmn.HexBytes `json:"root_hash"`

	Left  *pathWithNode `json:"left"`
	Right *pathWithNode `json:"right"`
}

func (proof *KeyAbsentProof) Root() []byte {
	return proof.RootHash
}

func (p *KeyAbsentProof) String() string {
	return fmt.Sprintf("KeyAbsentProof\nroot=%s\nleft=%s%#v\nright=%s%#v\n", p.RootHash, p.Left.Path, p.Left.Node, p.Right.Path, p.Right.Node)
}

// Verify verifies the proof is valid and returns an error if it isn't.
func (proof *KeyAbsentProof) Verify(key, value []byte, root []byte) error {
	if !bytes.Equal(proof.RootHash, root) {
		return errors.WithStack(ErrInvalidRoot)
	}
	if key == nil || value != nil {
		return ErrInvalidInputs
	}

	if proof.Left == nil && proof.Right == nil {
		return errors.WithStack(ErrInvalidProof)
	}
	if err := verifyPaths(proof.Left, proof.Right, key, key, root); err != nil {
		return err
	}

	return verifyKeyAbsence(proof.Left, proof.Right)
}

// Bytes returns a go-wire binary serialization
func (proof *KeyAbsentProof) Bytes() []byte {
	return append([]byte{keyAbsentMagicNumber}, wire.BinaryBytes(proof)...)
}

// readKeyAbsentProof will deserialize a KeyAbsentProof from bytes.
func readKeyAbsentProof(data []byte) (*KeyAbsentProof, error) {
	proof := new(KeyAbsentProof)
	err := wire.ReadBinaryBytes(data, &proof)
	return proof, err
}

// ReadKeyProof reads a KeyProof from a byte-slice.
func ReadKeyProof(data []byte) (KeyProof, error) {
	if len(data) == 0 {
		return nil, errors.New("proof bytes are empty")
	}
	b, val := data[0], data[1:]

	switch b {
	case keyExistsMagicNumber:
		return readKeyExistsProof(val)
	case keyAbsentMagicNumber:
		return readKeyAbsentProof(val)
	}
	return nil, errors.New("unrecognized proof")
}

///////////////////////////////////////////////////////////////////////////////

// InnerKeyProof represents a proof of existence of an inner node key.
type InnerKeyProof struct {
	*KeyExistsProof
}

// Verify verifies the proof is valid and returns an error if it isn't.
func (proof *InnerKeyProof) Verify(hash []byte, value []byte, root []byte) error {
	if !bytes.Equal(proof.RootHash, root) {
		return errors.WithStack(ErrInvalidRoot)
	}
	if hash == nil || value != nil {
		return errors.WithStack(ErrInvalidInputs)
	}
	return proof.PathToKey.verify(hash, root)
}

// ReadKeyInnerProof will deserialize a InnerKeyProof from bytes.
func ReadInnerKeyProof(data []byte) (*InnerKeyProof, error) {
	proof := new(InnerKeyProof)
	err := wire.ReadBinaryBytes(data, &proof)
	return proof, err
}
