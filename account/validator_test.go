package account

import (
	"testing"

	"github.com/hyperledger/burrow/permission"
	"github.com/stretchr/testify/assert"
)

func TestAlterPower(t *testing.T) {
	acc := NewAccountFromSecret("seeeeecret", permission.DefaultAccountPermissions)
	acc.AddToBalance(100)
	val := AsValidator(acc)
	valInc := val.WithNewPower(2442132)
	assert.Equal(t, uint64(100), val.Power())
	assert.Equal(t, uint64(2442132), valInc.Power())
}
