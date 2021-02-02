package state

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
)

func TestValidatorsReadWrite(t *testing.T) {
	s := NewState(dbm.NewMemDB())

	power := uint64(32432)
	v := validator.FromAccount(acm.NewAccountFromSecret("foobar"), power)

	_, _, err := s.Update(func(up Updatable) error {
		_, err := up.SetPower(v.GetPublicKey(), v.BigPower())
		return err
	})

	require.NoError(t, err)
	bigPower, err := s.Power(v.GetAddress())
	require.NoError(t, err)
	assert.Equal(t, power, bigPower.Uint64())

	fail := true
	err = s.IterateValidators(func(id crypto.Addressable, power *big.Int) error {
		fail = false
		assert.Equal(t, v.GetPublicKey(), id.GetPublicKey())
		assert.Equal(t, v.GetAddress(), id.GetAddress())
		assert.Equal(t, v.Power, power.Uint64())
		return nil
	})
	require.NoError(t, err)
	require.False(t, fail, "no validators in iteration")
}

func TestLoadValidatorRing(t *testing.T) {
	for commits := 1; commits < DefaultValidatorsWindowSize*7/2; commits++ {
		t.Run(fmt.Sprintf("TestLoadValidatorRing with %d commits", commits), func(t *testing.T) {
			testLoadValidatorRing(t, commits)
		})
	}
}

func testLoadValidatorRing(t *testing.T, commits int) {
	db := dbm.NewMemDB()
	s := NewState(db)

	var version int64
	var err error

	// we need to add a larger staked entity first
	// to prevent unbalancing the validator set
	_, err = s.writeState.SetPower(pub(0), pow(1000))
	require.NoError(t, err)
	_, version, err = s.commit()
	require.NoError(t, err)

	for i := 1; i <= commits; i++ {
		_, err = s.writeState.SetPower(pub(i), pow(i))
		require.NoError(t, err)
		_, version, err = s.commit()
		require.NoError(t, err)
	}

	ring := s.writeState.ring

	s = NewState(db)
	err = s.writeState.forest.Load(version)
	require.NoError(t, err)

	ringOut, err := LoadValidatorRing(version, DefaultValidatorsWindowSize, s.writeState.forest.GetImmutable)
	require.NoError(t, err)
	require.NoError(t, ring.Equal(ringOut))
}

func pow(p int) *big.Int {
	return big.NewInt(int64(p))
}

func pub(secret interface{}) *crypto.PublicKey {
	return acm.NewAccountFromSecret(fmt.Sprintf("%v", secret)).PublicKey
}
