package account

import (
	"encoding/binary"
	"fmt"

	"github.com/hyperledger/burrow/word"
	"golang.org/x/crypto/ripemd160"
)

type Address word.Word160

var ZeroAddress = Address{}

func AddressFromBytes(addr []byte) (address Address, err error) {
	if len(addr) != word.Word160Length {
		err = fmt.Errorf("slice passed as address '%X' has %d bytes but should "+
			"have %d bytes", addr, len(addr), word.Word160Length)
		// It is caller's responsibility to check for errors. If they ignore the error we'll assume they want the
		// best-effort mapping of the bytes passed to an  address so we don't return here
	}
	copy(address[:], addr)
	return address, nil
}

func AdddressFromString(str string) (address Address) {
	copy(address[:], ([]byte)(str))
	return
}

func MustAddressFromBytes(addr []byte) Address {
	address, err := AddressFromBytes(addr)
	if err != nil {
		panic(fmt.Errorf("error reading address from bytes that caller does not expect: %s", err))
	}
	return address
}

func AddressFromWord256(addr word.Word256) Address {
	return Address(addr.Word160())
}

func (address Address) Word256() word.Word256 {
	return word.Word160(address).Word256()
}

// Copy address and return a slice onto the copy
func (address Address) Bytes() []byte {
	addressCopy := address
	return addressCopy[:]
}

func (address Address) String() string {
	return fmt.Sprintf("%X", address[:])
}

func NewContractAddress(caller Address, sequence int64) (newAddr Address) {
	temp := make([]byte, 32+8)
	copy(temp, caller[:])
	binary.BigEndian.PutUint64(temp[32:], uint64(sequence))
	hasher := ripemd160.New()
	hasher.Write(temp) // does not error
	copy(newAddr[:], hasher.Sum(nil))
	return
}
