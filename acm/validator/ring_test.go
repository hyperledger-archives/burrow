package validator

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var pubA = pubKey(1)
var pubB = pubKey(2)
var pubC = pubKey(3)

func TestValidatorsWindow_AlterPower(t *testing.T) {
	vsBase := NewSet()
	powAInitial := int64(10000)
	vsBase.ChangePower(pubA, big.NewInt(powAInitial))

	vs := Copy(vsBase)
	vw := NewRing(vs, 3)

	// Just allowable validator tide
	var powA, powB, powC int64 = 7000, 23, 309
	powerChange, totalFlow, err := alterPowers(t, vw, powA, powB, powC)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(powA+powB+powC-powAInitial), powerChange)
	assert.Equal(t, big.NewInt(powAInitial/3-1), totalFlow)

	// This one is not
	vs = Copy(vsBase)
	vw = NewRing(vs, 5)
	powA, powB, powC = 7000, 23, 310
	powerChange, totalFlow, err = alterPowers(t, vw, powA, powB, powC)
	require.Error(t, err)

	powA, powB, powC = 7000, 23, 309
	powerChange, totalFlow, err = alterPowers(t, vw, powA, powB, powC)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(powA+powB+powC-powAInitial), powerChange)
	assert.Equal(t, big.NewInt(powAInitial/3-1), totalFlow)

	powA, powB, powC = 7000, 23, 309
	powerChange, totalFlow, err = alterPowers(t, vw, powA, powB, powC)
	require.NoError(t, err)
	assertZero(t, powerChange)
	assertZero(t, totalFlow)

	_, err = vw.AlterPower(pubA, big.NewInt(8000))
	assert.NoError(t, err)

	// Should fail - not enough flow left
	_, err = vw.AlterPower(pubB, big.NewInt(2000))
	assert.Error(t, err)

	// Take a bit off shouhd work
	_, err = vw.AlterPower(pubA, big.NewInt(7000))
	assert.NoError(t, err)

	_, err = vw.AlterPower(pubB, big.NewInt(2000))
	assert.NoError(t, err)
	vw.Rotate()

	powerChange, totalFlow, err = alterPowers(t, vw, powA, powB, powC)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(-1977), powerChange)
	assert.Equal(t, big.NewInt(1977), totalFlow)

	powerChange, totalFlow, err = alterPowers(t, vw, powA, powB, powC)
	require.NoError(t, err)
	assertZero(t, powerChange)
	assert.Equal(t, big0, totalFlow)

	powerChange, totalFlow, err = alterPowers(t, vw, powA, powB, powC)
	require.NoError(t, err)
	assertZero(t, powerChange)
	assert.Equal(t, big0, totalFlow)
}

func TestValidatorsRing_Persistable(t *testing.T) {

	vs := NewSet()
	powAInitial := int64(10000)
	vs.ChangePower(pubA, big.NewInt(powAInitial))
	vw := NewRing(vs, 30)

	for i := int64(0); i < 61; i++ {
		_, _, err := alterPowers(t, vw, 10000, 200*i, 200*((i+1)%4))
		require.NoError(t, err)
	}

	vwOut := UnpersistRing(vw.Persistable())
	assert.True(t, vw.Equal(vwOut), "should re equal across persistence")
}

func alterPowers(t testing.TB, vw *Ring, powA, powB, powC int64) (powerChange, totalFlow *big.Int, err error) {
	fmt.Println(vw)
	_, err = vw.AlterPower(pubA, big.NewInt(powA))
	if err != nil {
		return nil, nil, err
	}
	_, err = vw.AlterPower(pubB, big.NewInt(powB))
	if err != nil {
		return nil, nil, err
	}
	_, err = vw.AlterPower(pubC, big.NewInt(powC))
	if err != nil {
		return nil, nil, err
	}
	maxFlow := vw.MaxFlow()
	powerChange, totalFlow, err = vw.Rotate()
	require.NoError(t, err)
	// totalFlow > maxFlow
	if totalFlow.Cmp(maxFlow) == 1 {
		return powerChange, totalFlow, fmt.Errorf("totalFlow (%v) exceeds maxFlow (%v)", totalFlow, maxFlow)
	}

	return powerChange, totalFlow, nil
}

// Since we have -0 and 0 with big.Int due to its representation with a neg flag
func assertZero(t testing.TB, i *big.Int) {
	assert.True(t, big0.Cmp(i) == 0, "expected 0 but got %v", i)
}
