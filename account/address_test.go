package account

import (
	"testing"

	"encoding/json"

	"github.com/stretchr/testify/assert"
)

func TestNewContractAddress(t *testing.T) {
	addr := NewContractAddress(Address{
		233, 181, 216, 115, 19,
		53, 100, 101, 250, 227,
		60, 64, 108, 226, 194,
		151, 157, 230, 11, 203,
	}, 1)

	assert.Equal(t, Address{
		73, 234, 48, 252, 174,
		115, 27, 222, 54, 116,
		47, 133, 144, 21, 73,
		245, 21, 234, 26, 50,
	}, addr)
}

func TestAddress_MarshalJSON(t *testing.T) {
	addr := Address{
		73, 234, 48, 252, 174,
		115, 27, 222, 54, 116,
		47, 133, 144, 21, 73,
		245, 21, 234, 26, 50,
	}

	bs, err := json.Marshal(addr)
	assert.NoError(t, err)

	addrOut := new(Address)
	err = json.Unmarshal(bs, addrOut)

	assert.Equal(t, addr, *addrOut)
}
