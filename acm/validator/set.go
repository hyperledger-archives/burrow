package validator

import (
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/hyperledger/burrow/crypto"
)

var big0 = big.NewInt(0)

// A Validator multiset - can be used to capture the global state of validators or as an accumulator each block
type Set struct {
	powers     map[crypto.Address]*big.Int
	publicKeys map[crypto.Address]crypto.Addressable
	totalPower *big.Int
	trim       bool
}

func newSet() *Set {
	return &Set{
		totalPower: new(big.Int),
		powers:     make(map[crypto.Address]*big.Int),
		publicKeys: make(map[crypto.Address]crypto.Addressable),
	}
}

// Create a new Validators which can act as an accumulator for validator power changes
func NewSet() *Set {
	return newSet()
}

// Like Set but removes entries when power is set to 0 this make Count() == CountNonZero() and prevents a set from leaking
// but does mean that a zero will not be iterated over when performing an update which is necessary in Ring
func NewTrimSet() *Set {
	s := newSet()
	s.trim = true
	return s
}

// Implements Writer, but will never error
func (vs *Set) AlterPower(id crypto.PublicKey, power *big.Int) (flow *big.Int, err error) {
	return vs.ChangePower(id, power), nil
}

// Add the power of a validator and returns the flow into that validator
func (vs *Set) ChangePower(id crypto.PublicKey, power *big.Int) *big.Int {
	address := id.Address()
	// Calculcate flow into this validator (postive means in, negative means out)
	flow := new(big.Int).Sub(power, vs.Power(id.Address()))
	vs.totalPower.Add(vs.totalPower, flow)

	if vs.trim && power.Sign() == 0 {
		delete(vs.publicKeys, address)
		delete(vs.powers, address)
	} else {
		vs.publicKeys[address] = crypto.NewAddressable(id)
		vs.powers[address] = new(big.Int).Set(power)
	}
	return flow
}

func (vs *Set) TotalPower() *big.Int {
	return new(big.Int).Set(vs.totalPower)
}

// Returns the power of id but only if it is set
func (vs *Set) MaybePower(id crypto.Address) *big.Int {
	if vs.powers[id] == nil {
		return nil
	}
	return new(big.Int).Set(vs.powers[id])
}

func (vs *Set) Power(id crypto.Address) *big.Int {
	if vs.powers[id] == nil {
		return new(big.Int)
	}
	return new(big.Int).Set(vs.powers[id])
}

func (vs *Set) Equal(vsOther *Set) bool {
	if vs.Count() != vsOther.Count() {
		return false
	}
	// Stop iteration IFF we find a non-matching validator
	return !vs.Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		otherPower := vsOther.Power(id.Address())
		if otherPower.Cmp(power) != 0 {
			return true
		}
		return false
	})
}

// Iterates over validators sorted by address
func (vs *Set) Iterate(iter func(id crypto.Addressable, power *big.Int) (stop bool)) (stopped bool) {
	if vs == nil {
		return
	}
	addresses := make(crypto.Addresses, 0, len(vs.powers))
	for address := range vs.powers {
		addresses = append(addresses, address)
	}
	sort.Sort(addresses)
	for _, address := range addresses {
		if iter(vs.publicKeys[address], new(big.Int).Set(vs.powers[address])) {
			return true
		}
	}
	return
}

func (vs *Set) CountNonZero() int {
	var count int
	vs.Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		if power.Sign() != 0 {
			count++
		}
		return
	})
	return count
}

func (vs *Set) Count() int {
	return len(vs.publicKeys)
}

func (vs *Set) Validators() []*Validator {
	if vs == nil {
		return nil
	}
	pvs := make([]*Validator, 0, vs.Count())
	vs.Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		pvs = append(pvs, &Validator{PublicKey: id.PublicKey(), Power: power.Uint64()})
		return
	})
	return pvs
}

func UnpersistSet(pvs []*Validator) *Set {
	vs := NewSet()
	for _, pv := range pvs {
		vs.ChangePower(pv.PublicKey, new(big.Int).SetUint64(pv.Power))
	}
	return vs
}

func (vs *Set) String() string {
	return fmt.Sprintf("Validators{TotalPower: %v; Count: %v; %v}", vs.TotalPower(), vs.Count(),
		vs.Strings())
}

func (vs *Set) Strings() string {
	strs := make([]string, 0, vs.Count())
	vs.Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		strs = append(strs, fmt.Sprintf("%v->%v", id.Address(), power))
		return
	})
	return strings.Join(strs, ", ")
}
