package validator

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/stretchr/testify/assert"
)

func TestValidators_AlterPower(t *testing.T) {
	vs := NewSet()
	pow1 := big.NewInt(2312312321)
	pubA := pubKey(1)
	vs.ChangePower(pubA, pow1)
	assert.Equal(t, pow1, vs.TotalPower())
	vs.ChangePower(pubA, big.NewInt(0))
	assertZero(t, vs.TotalPower())
}

func pubKey(secret interface{}) crypto.PublicKey {
	return acm.NewConcreteAccountFromSecret(fmt.Sprintf("%v", secret)).PublicKey
}
