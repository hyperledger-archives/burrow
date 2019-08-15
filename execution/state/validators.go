package state

import (
	"math/big"

	"github.com/hyperledger/burrow/encoding"
	"github.com/hyperledger/burrow/genesis"

	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/storage"

	"github.com/hyperledger/burrow/crypto"
)

// Initialises the validator Ring from the validator storage in forest
func LoadValidatorRing(version int64, ringSize int,
	getImmutable func(version int64) (*storage.ImmutableForest, error)) (*validator.Ring, error) {

	// In this method we have to page through previous version of the tree in order to reconstruct the in-memory
	// ring structure. The corner cases are a little subtle but printing the buckets helps

	// The basic idea is to load each version of the tree ringSize back, work out the difference that must have occurred
	// between each bucket in the ring, and apply each diff to the ring. Once the ring is full it is symmetrical (up to
	// a reindexing). If we are loading a chain whose height is less than the ring size we need to get the initial state
	// correct

	startVersion := version - int64(ringSize)
	if startVersion < 1 {
		// The ring will not be fully populated
		startVersion = 1
	}
	var err error
	// Read state to pull immutable forests from
	rs := &ReadState{}
	// Start with an empty ring - we want the initial bucket to have no cumulative power
	ring := validator.NewRing(nil, ringSize)
	// Load the IAVL state
	rs.Forest, err = getImmutable(startVersion)
	if err != nil {
		return nil, err
	}
	// Write the validator state at startVersion from IAVL tree into the ring's current bucket delta
	err = validator.Write(ring, rs)
	if err != nil {
		return nil, err
	}
	// Rotate, now we have [ {bucket 0: cum: {}, delta: {start version changes} }, {bucket 1: cum: {start version changes}, delta {}, ... ]
	// which is what we need (in particular we need this initial state if we are loading into a incompletely populated ring
	_, _, err = ring.Rotate()
	if err != nil {
		return nil, err
	}

	// Rebuild validator Ring
	for v := startVersion + 1; v <= version; v++ {
		// Update IAVL read state to version of interest
		rs.Forest, err = getImmutable(v)
		if err != nil {
			return nil, err
		}
		// Calculate the difference between the rings current cum and what is in state at this version
		diff, err := validator.Diff(ring.CurrentSet(), rs)
		if err != nil {
			return nil, err
		}
		// Write that diff into the ring (just like it was when it was originally written to setPower)
		err = validator.Write(ring, diff)
		if err != nil {
			return nil, err
		}
		// Rotate just like it was on the original commit
		_, _, err = ring.Rotate()
		if err != nil {
			return nil, err
		}
	}
	// Our ring should be the same up to symmetry in its index so we reindex to regain equality with the version we are loading
	// This is the head index we would have had if we had started from version 1 like the chain did
	ring.ReIndex(int(version % int64(ringSize)))
	return ring, err
}

func (ws *writeState) MakeGenesisValidators(genesisDoc *genesis.GenesisDoc) error {
	for _, gv := range genesisDoc.Validators {
		_, err := ws.SetPower(gv.PublicKey, new(big.Int).SetUint64(gv.Amount))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ReadState) Power(id crypto.Address) (*big.Int, error) {
	tree, err := s.Forest.Reader(keys.Validator.Prefix())
	if err != nil {
		return nil, err
	}
	bs := tree.Get(keys.Validator.KeyNoPrefix(id))
	if len(bs) == 0 {
		return new(big.Int), nil
	}
	v := new(validator.Validator)
	err = encoding.Decode(bs, v)
	if err != nil {
		return nil, err
	}
	return v.BigPower(), nil
}

func (s *ReadState) IterateValidators(fn func(id crypto.Addressable, power *big.Int) error) error {
	tree, err := s.Forest.Reader(keys.Validator.Prefix())
	if err != nil {
		return err
	}
	return tree.Iterate(nil, nil, true, func(_, value []byte) error {
		v := new(validator.Validator)
		err = encoding.Decode(value, v)
		if err != nil {
			return err
		}
		return fn(v, v.BigPower())
	})
}

func (ws *writeState) SetPower(id crypto.PublicKey, power *big.Int) (*big.Int, error) {
	// SetPower in ring
	flow, err := ws.ring.SetPower(id, power)
	if err != nil {
		return nil, err
	}
	// Set power in versioned state
	return flow, ws.setPower(id, power)
}

func (ws *writeState) setPower(id crypto.PublicKey, power *big.Int) error {
	tree, err := ws.forest.Writer(keys.Validator.Prefix())
	if err != nil {
		return err
	}
	key := keys.Validator.KeyNoPrefix(id.GetAddress())
	if power.Sign() == 0 {
		tree.Delete(key)
		return nil
	}
	bs, err := encoding.Encode(validator.New(id, power))
	if err != nil {
		return err
	}
	tree.Set(key, bs)
	return nil
}
