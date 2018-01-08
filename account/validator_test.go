package account

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlterPower(t *testing.T) {
	val := AsValidator(NewConcreteAccountFromSecret("seeeeecret").Account())
	valInc := val.WithNewPower(2442132)
	assert.Equal(t, uint64(0), val.Power())
	assert.Equal(t, uint64(2442132), valInc.Power())
}
