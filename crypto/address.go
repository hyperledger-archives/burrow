package crypto

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/burrow/binary"
	"github.com/tmthrgd/go-hex"
	"golang.org/x/crypto/ripemd160"
)

type Addressable interface {
	// Get the 20 byte EVM address of this account
	Address() Address
	// Public key from which the Address is derived
	PublicKey() PublicKey
}

func NewAddressable(publicKey PublicKey) Addressable {
	return &memoizedAddressable{
		address:   publicKey.Address(),
		publicKey: publicKey,
	}
}

type memoizedAddressable struct {
	publicKey PublicKey
	address   Address
}

func MemoizeAddressable(addressable Addressable) Addressable {
	if a, ok := addressable.(*memoizedAddressable); ok {
		return a
	}
	return NewAddressable(addressable.PublicKey())
}

func (a *memoizedAddressable) PublicKey() PublicKey {
	return a.publicKey
}

func (a *memoizedAddressable) Address() Address {
	return a.address
}

type Address binary.Word160

type Addresses []Address

func (as Addresses) Len() int {
	return len(as)
}

func (as Addresses) Less(i, j int) bool {
	return bytes.Compare(as[i][:], as[j][:]) < 0
}
func (as Addresses) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}

const AddressLength = binary.Word160Length
const AddressHexLength = 2 * AddressLength

var ZeroAddress = Address{}

// Returns a pointer to an Address that is nil iff len(bs) == 0 otherwise does the same as AddressFromBytes
func MaybeAddressFromBytes(bs []byte) (*Address, error) {
	if len(bs) == 0 {
		return nil, nil
	}
	address, err := AddressFromBytes(bs)
	if err != nil {
		return nil, err
	}
	return &address, nil
}

// Returns an address consisting of the first 20 bytes of bs, return an error if the bs does not have length exactly 20
// but will still return either: the bytes in bs padded on the right or the first 20 bytes of bs truncated in any case.
func AddressFromBytes(bs []byte) (address Address, err error) {
	if len(bs) != binary.Word160Length {
		err = fmt.Errorf("slice passed as address '%X' has %d bytes but should have %d bytes",
			bs, len(bs), binary.Word160Length)
		// It is caller's responsibility to check for errors. If they ignore the error we'll assume they want the
		// best-effort mapping of the bytes passed to an address so we don't return here
	}
	copy(address[:], bs)
	return
}

func AddressFromHexString(str string) (Address, error) {
	bs, err := hex.DecodeString(str)
	if err != nil {
		return ZeroAddress, err
	}
	return AddressFromBytes(bs)
}

func MustAddressFromBytes(addr []byte) Address {
	address, err := AddressFromBytes(addr)
	if err != nil {
		panic(fmt.Errorf("error reading address from bytes that caller does not expect: %s", err))
	}
	return address
}

func AddressFromWord256(addr binary.Word256) Address {
	return Address(addr.Word160())
}

func (address Address) Word256() binary.Word256 {
	return binary.Word160(address).Word256()
}

// Copy address and return a slice onto the copy
func (address Address) Bytes() []byte {
	addressCopy := address
	return addressCopy[:]
}

func (address Address) String() string {
	return hex.EncodeUpperToString(address[:])
}

func (address *Address) UnmarshalJSON(data []byte) error {
	str := new(string)
	err := json.Unmarshal(data, str)
	if err != nil {
		return err
	}
	err = address.UnmarshalText([]byte(*str))
	if err != nil {
		return err
	}
	return nil
}

func (address Address) MarshalJSON() ([]byte, error) {
	text, err := address.MarshalText()
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(text))
}

func (address *Address) UnmarshalText(text []byte) error {
	if len(text) != AddressHexLength {
		return fmt.Errorf("address hex '%s' has length %v but must have length %v to be a valid address",
			string(text), len(text), AddressHexLength)
	}
	_, err := hex.Decode(address[:], text)
	return err
}

func (address Address) MarshalText() ([]byte, error) {
	return ([]byte)(hex.EncodeUpperToString(address[:])), nil

}

// Gogo proto support
func (address *Address) Marshal() ([]byte, error) {
	if address == nil {
		return nil, nil
	}
	return address.Bytes(), nil
}

func (address *Address) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if len(data) != binary.Word160Length {
		return fmt.Errorf("error unmarshallling address '%X' from bytes: %d bytes but should have %d bytes",
			data, len(data), binary.Word160Length)
	}
	copy(address[:], data)
	return nil
}

func (address *Address) MarshalTo(data []byte) (int, error) {
	return copy(data, address[:]), nil
}

func (address *Address) Size() int {
	return binary.Word160Length
}

func NewContractAddress(caller Address, sequence uint64) (newAddr Address) {
	temp := make([]byte, 32+8)
	copy(temp, caller[:])
	binary.PutUint64BE(temp[32:], uint64(sequence))
	hasher := ripemd160.New()
	hasher.Write(temp) // does not error
	copy(newAddr[:], hasher.Sum(nil))
	return
}
