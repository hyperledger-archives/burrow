package crypto

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContractAddress(t *testing.T) {
	addr := NewContractAddress(Address{
		233, 181, 216, 115, 19,
		53, 100, 101, 250, 227,
		60, 64, 108, 226, 194,
		151, 157, 230, 11, 203,
	}, []byte{1})

	assert.Equal(t, Address{
		0xe2, 0x9d, 0x64, 0x7d, 0xce, 0xd6, 0x48, 0x12, 0xd4, 0x44,
		0xa6, 0x7d, 0x80, 0x5, 0xcd, 0x6, 0x9, 0xb2, 0x2f, 0x49,
	}, addr)
}

func TestNewContractAddress2(t *testing.T) {
	addr := NewContractAddress2(Address{
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
	}, [32]byte{
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
	}, []byte{0x00})

	assert.Equal(t, Address{
		77, 26, 46, 43, 180,
		248, 143, 2, 80, 242,
		111, 255, 240, 152, 176,
		179, 11, 38, 191, 56,
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
	require.NoError(t, err)

	assert.Equal(t, addr, *addrOut)
}

func TestAddress_MarshalText(t *testing.T) {
	addr := Address{
		73, 234, 48, 252, 174,
		115, 27, 222, 54, 116,
		47, 133, 144, 21, 73,
		245, 21, 234, 26, 50,
	}

	bs, err := addr.MarshalText()
	assert.NoError(t, err)

	addrOut := new(Address)
	err = addrOut.UnmarshalText(bs)
	require.NoError(t, err)

	assert.Equal(t, addr, *addrOut)
}

func TestAddress_Length(t *testing.T) {
	addrOut := new(Address)
	err := addrOut.UnmarshalText(([]byte)("49EA30FCAE731BDE36742F85901549F515EA1A10"))
	require.NoError(t, err)

	err = addrOut.UnmarshalText(([]byte)("49EA30FCAE731BDE36742F85901549F515EA1A1"))
	assert.Error(t, err, "address too short")

	err = addrOut.UnmarshalText(([]byte)("49EA30FCAE731BDE36742F85901549F515EA1A1020"))
	assert.Error(t, err, "address too long")
}

func TestAddress_Sort(t *testing.T) {
	addresses := Addresses{
		{2, 3, 4},
		{3, 1, 2},
		{2, 1, 2},
	}
	sorted := make(Addresses, len(addresses))
	copy(sorted, addresses)
	sort.Stable(sorted)
	assert.Equal(t, Addresses{
		{2, 1, 2},
		{2, 3, 4},
		{3, 1, 2},
	}, sorted)
}

func TestSequenceNonce(t *testing.T) {
	nonce := SequenceNonce(Address{'w', 'h', 'a', 't', 'n', 'o', 'n', 'c', 'e', 'n', 's', 'e'}, 36345634534)
	assert.Equal(t, []byte{
		0x39, 0x66, 0x0d, 0x93, 0x12, 0xb5, 0x48, 0xed,
		0xc6, 0xb7, 0x46, 0x03, 0xad, 0xc3, 0xe6, 0xe2,
		0x3f, 0x6b, 0x35, 0x52, 0x54, 0x96, 0x30, 0x3e,
		0x18, 0x94, 0xc0, 0x78, 0xb5, 0x33, 0x73, 0xb5,
	}, nonce)
}
