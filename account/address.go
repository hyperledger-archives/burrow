package account

import (
	"fmt"

	"github.com/hyperledger/burrow/word"
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

func (address Address) Bytes() []byte {
	addressCopy := address
	return addressCopy[:]
}

func (address Address) String() string {
	return fmt.Sprintf("%X", address[:])
}
