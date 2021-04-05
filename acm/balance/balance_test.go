package balance

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSum(t *testing.T) {
	one := New().Power(23223).Native(34).Native(1111)
	two := New().Power(3).Native(22)
	sum := one.Sum(two)
	assert.Equal(t, New().Power(23226).Native(1167).Sort(), sum)
}

func TestSort(t *testing.T) {
	balances := New().Power(232).Native(2523543).Native(232).Power(2).Power(4).Native(1)
	sortedBalances := New().Native(1).Native(232).Native(2523543).Power(2).Power(4).Power(232)
	sort.Sort(balances)
	assert.Equal(t, sortedBalances, balances)
}

func TestEtherConversion(t *testing.T) {
	wei := NativeToWei(1)
	assert.Equal(t, wei.String(), "1000000000000000000", "must equal one ether")
	native := WeiToNative(wei)
	assert.Equal(t, native.Uint64(), uint64(1))
}
