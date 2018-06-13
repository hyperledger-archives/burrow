package blockchain

import (
	"fmt"

	"sort"

	"bytes"
	"encoding/binary"

	burrowBinary "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
)

// A Validator multiset
type Validators struct {
	power      map[crypto.Address]uint64
	publicKey  map[crypto.Address]crypto.PublicKey
	totalPower uint64
}

func NewValidators() Validators {
	return Validators{
		power:     make(map[crypto.Address]uint64),
		publicKey: make(map[crypto.Address]crypto.PublicKey),
	}
}

// Add the power of a validator
func (vs *Validators) AlterPower(publicKey crypto.PublicKey, power uint64) error {
	address := publicKey.Address()
	// Remove existing power (possibly 0) from total
	vs.totalPower -= vs.power[address]
	if burrowBinary.IsUint64SumOverflow(vs.totalPower, power) {
		// Undo removing existing power
		vs.totalPower += vs.power[address]
		return fmt.Errorf("could not increase total validator power by %v from %v since that would overflow "+
			"uint64", power, vs.totalPower)
	}
	vs.publicKey[address] = publicKey
	vs.power[address] = power
	// Note we are adjusting by the difference in power (+/-) since we subtracted the previous amount above
	vs.totalPower += power
	return nil
}

func (vs *Validators) AddPower(publicKey crypto.PublicKey, power uint64) error {
	currentPower := vs.power[publicKey.Address()]
	if burrowBinary.IsUint64SumOverflow(currentPower, power) {
		return fmt.Errorf("could add power %v to validator %v with power %v because that would overflow uint64",
			power, publicKey.Address(), currentPower)
	}
	return vs.AlterPower(publicKey, vs.power[publicKey.Address()]+power)
}

func (vs *Validators) SubtractPower(publicKey crypto.PublicKey, power uint64) error {
	currentPower := vs.power[publicKey.Address()]
	if currentPower < power {
		return fmt.Errorf("could subtract power %v from validator %v with power %v because that would "+
			"underflow uint64", power, publicKey.Address(), currentPower)
	}
	return vs.AlterPower(publicKey, vs.power[publicKey.Address()]-power)
}

// Iterates over validators sorted by address
func (vs *Validators) Iterate(iter func(publicKey crypto.PublicKey, power uint64) (stop bool)) (stopped bool) {
	addresses := make(crypto.Addresses, 0, len(vs.power))
	for address := range vs.power {
		addresses = append(addresses, address)
	}
	sort.Sort(addresses)
	for _, address := range addresses {
		if iter(vs.publicKey[address], vs.power[address]) {
			return true
		}
	}
	return false
}

func (vs *Validators) Length() int {
	return len(vs.power)
}

func (vs *Validators) TotalPower() uint64 {
	return vs.totalPower
}

// Uses the fixed width public key encoding to
func (vs *Validators) Encode() []byte {
	buffer := new(bytes.Buffer)
	// varint buffer
	buf := make([]byte, 8)
	vs.Iterate(func(publicKey crypto.PublicKey, power uint64) (stop bool) {
		buffer.Write(publicKey.Encode())
		buffer.Write(buf[:binary.PutUvarint(buf, power)])
		return
	})
	return buffer.Bytes()
}

// Decodes validators encoded with Encode - expects the exact encoded size with no trailing bytes
func DecodeValidators(encoded []byte, validators *Validators) error {
	publicKey := new(crypto.PublicKey)
	i := 0
	for i < len(encoded) {
		n, err := crypto.DecodePublicKeyFixedWidth(encoded[i:], publicKey)
		if err != nil {
			return err
		}
		i += n
		power, n := binary.Uvarint(encoded[i:])
		if n <= 0 {
			return fmt.Errorf("error decoding uint64 from validators binary encoding")
		}
		i += n
		err = validators.AlterPower(*publicKey, power)
		if err != nil {
			return err
		}
	}
	return nil
}
