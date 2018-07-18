package blockchain

import (
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/hyperledger/burrow/crypto"
)

var big0 = big.NewInt(0)

// A Validator multiset - can be used to capture the global state of validators or as an accumulator each block
type Validators struct {
	powers     map[crypto.Address]*big.Int
	publicKeys map[crypto.Address]crypto.Addressable
	totalPower *big.Int
}

type ValidatorSet interface {
	AlterPower(id crypto.Addressable, power *big.Int) (flow *big.Int, err error)
}

// Create a new Validators which can act as an accumulator for validator power changes
func NewValidators() *Validators {
	return &Validators{
		totalPower: new(big.Int),
		powers:     make(map[crypto.Address]*big.Int),
		publicKeys: make(map[crypto.Address]crypto.Addressable),
	}
}

// Add the power of a validator and returns the flow into that validator
func (vs *Validators) AlterPower(id crypto.Addressable, power *big.Int) *big.Int {
	if power.Sign() == -1 {
		panic("ASRRRH")
	}
	address := id.Address()
	// Calculcate flow into this validator (postive means in, negative means out)
	flow := new(big.Int).Sub(power, vs.Power(id))
	vs.totalPower.Add(vs.totalPower, flow)
	if power.Cmp(big0) == 0 {
		// Remove from set so that we return an accurate length
		delete(vs.publicKeys, address)
		delete(vs.powers, address)
		return flow
	}
	vs.publicKeys[address] = crypto.MemoizeAddressable(id)
	vs.powers[address] = new(big.Int).Set(power)
	return flow
}

// Adds vsOther to vs
func (vs *Validators) Add(vsOther *Validators) {
	vsOther.Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		vs.AddPower(id, power)
		return
	})
}

func (vs *Validators) AddPower(id crypto.Addressable, power *big.Int) {
	// Current power + power
	vs.AlterPower(id, new(big.Int).Add(vs.Power(id), power))
}

// Subtracts vsOther from vs
func (vs *Validators) Subtract(vsOther *Validators) {
	vsOther.Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		vs.SubtractPower(id, power)
		return
	})
}

func (vs *Validators) SubtractPower(id crypto.Addressable, power *big.Int) {
	// Current power - power
	thisPower := vs.Power(id)
	vs.AlterPower(id, new(big.Int).Sub(thisPower, power))
}

func (vs *Validators) Power(id crypto.Addressable) *big.Int {
	if vs.powers[id.Address()] == nil {
		return new(big.Int)
	}
	return new(big.Int).Set(vs.powers[id.Address()])
}

func (vs *Validators) Equal(vsOther *Validators) bool {
	if vs.Count() != vsOther.Count() {
		return false
	}
	// Stop iteration IFF we find a non-matching validator
	return !vs.Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		otherPower := vsOther.Power(id)
		if otherPower.Cmp(power) != 0 {
			return true
		}
		return false
	})
}

// Iterates over validators sorted by address
func (vs *Validators) Iterate(iter func(id crypto.Addressable, power *big.Int) (stop bool)) (stopped bool) {
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

func (vs *Validators) Count() int {
	return len(vs.publicKeys)
}

func (vs *Validators) TotalPower() *big.Int {
	return new(big.Int).Set(vs.totalPower)
}

func (vs *Validators) Copy() *Validators {
	vsCopy := NewValidators()
	vs.Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		vsCopy.AlterPower(id, power)
		return
	})
	return vsCopy
}

type PersistedValidator struct {
	PublicKey  crypto.PublicKey
	PowerBytes []byte
}

func (vs *Validators) Persistable() []PersistedValidator {
	if vs == nil {
		return nil
	}
	pvs := make([]PersistedValidator, 0, vs.Count())
	vs.Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		pvs = append(pvs, PersistedValidator{PublicKey: id.PublicKey(), PowerBytes: power.Bytes()})
		return
	})
	return pvs
}

func UnpersistValidators(pvs []PersistedValidator) *Validators {
	vs := NewValidators()
	for _, pv := range pvs {
		power := new(big.Int).SetBytes(pv.PowerBytes)
		vs.AlterPower(pv.PublicKey, power)
	}
	return vs
}

func (vs *Validators) String() string {
	return fmt.Sprintf("Validators{TotalPower: %v; Count: %v; %v}", vs.TotalPower(), vs.Count(),
		vs.ValidatorStrings())
}

func (vs *Validators) ValidatorStrings() string {
	strs := make([]string, 0, vs.Count())
	vs.Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		strs = append(strs, fmt.Sprintf("%v->%v", id.Address(), power))
		return
	})
	return strings.Join(strs, ", ")
}
