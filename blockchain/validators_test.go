package blockchain

import (
	"testing"

	"github.com/hyperledger/burrow/permission"

	"fmt"

	"math/rand"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidators_AlterPower(t *testing.T) {
	vs := NewValidators()
	pow1 := uint64(2312312321)
	assert.NoError(t, vs.AlterPower(pubKey(1), pow1))
	assert.Equal(t, pow1, vs.TotalPower())
}

func TestValidators_Encode(t *testing.T) {
	vs := NewValidators()
	rnd := rand.New(rand.NewSource(43534543))
	for i := 0; i < 100; i++ {
		power := uint64(rnd.Intn(10))
		require.NoError(t, vs.AlterPower(pubKey(rnd.Int63()), power))
	}
	encoded := vs.Encode()
	vsOut := NewValidators()
	require.NoError(t, DecodeValidators(encoded, &vsOut))
	// Check decoded matches encoded
	var publicKeyPower []interface{}
	vs.Iterate(func(publicKey crypto.PublicKey, power uint64) (stop bool) {
		publicKeyPower = append(publicKeyPower, publicKey, power)
		return
	})
	vsOut.Iterate(func(publicKey crypto.PublicKey, power uint64) (stop bool) {
		assert.Equal(t, publicKeyPower[0], publicKey)
		assert.Equal(t, publicKeyPower[1], power)
		publicKeyPower = publicKeyPower[2:]
		return
	})
	assert.Len(t, publicKeyPower, 0, "should exhaust all validators in decoded multiset")
}

func pubKey(secret interface{}) crypto.PublicKey {
	return acm.NewAccountFromSecret(fmt.Sprintf("%v", secret), permission.ZeroAccountPermissions).PublicKey()
}
