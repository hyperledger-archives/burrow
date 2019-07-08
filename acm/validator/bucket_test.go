package validator

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

var pubA = pubKey(1)
var pubB = pubKey(2)
var pubC = pubKey(3)

func TestBucket_SetPower(t *testing.T) {
	base := NewBucket()
	_, err := base.SetPower(pubA, new(big.Int).Sub(maxTotalVotingPower, big3))
	require.NoError(t, err)

	bucket := NewBucket(base.Next)

	flow, err := bucket.SetPower(pubA, new(big.Int).Sub(maxTotalVotingPower, big2))
	require.NoError(t, err)
	require.Equal(t, big1.Int64(), flow.Int64())

	flow, err = bucket.SetPower(pubA, new(big.Int).Sub(maxTotalVotingPower, big1))
	require.NoError(t, err)
	require.Equal(t, big2.Int64(), flow.Int64())

	flow, err = bucket.SetPower(pubA, maxTotalVotingPower)
	require.NoError(t, err)
	require.Equal(t, big3.Int64(), flow.Int64())

	_, err = bucket.SetPower(pubA, new(big.Int).Add(maxTotalVotingPower, big1))
	require.Error(t, err, "should fail as we would breach total power")

	_, err = bucket.SetPower(pubB, big1)
	require.Error(t, err, "should fail as we would breach total power")

	// Drop A and raise B - should now succeed
	flow, err = bucket.SetPower(pubA, new(big.Int).Sub(maxTotalVotingPower, big1))
	require.NoError(t, err)
	require.Equal(t, big2.Int64(), flow.Int64())

	flow, err = bucket.SetPower(pubB, big1)
	require.NoError(t, err)
	require.Equal(t, big1.Int64(), flow.Int64())
}
