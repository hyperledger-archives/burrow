package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"

	acm "github.com/hyperledger/burrow/account"
	rpc_core "github.com/tendermint/tendermint/rpc/core"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tm_types "github.com/tendermint/tendermint/types"
)

type ValidatorSet interface {
	TotalPower() int
	SetMaximumPower(maximumPower int)
	MaximumPower() int
	AdjustPower(height int64) error
	Validators() []acm.Validator
	SetLeavers() []acm.Validator
	JoinToTheSet(validator acm.Validator) error
	LeaveFromTheSet(validator acm.Validator) error
	IsValidatorInSet(address acm.Address) bool
	JSONBytes() ([]byte, error)
}

type validatorListProxy interface {
	Validators(height int64) (*ctypes.ResultValidators, error)
}

type _validatorListProxy struct{}

func (vlp _validatorListProxy) Validators(height int64) (*ctypes.ResultValidators, error) {
	result, err := rpc_core.Validators(&height)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type validatorSet struct {
	maximumPower int
	validators   []acm.Validator
	setLeavers   []acm.Validator
	proxy        validatorListProxy
}

func newValidatorSet(maximumPower int, validators []acm.Validator) *validatorSet {
	set := &validatorSet{
		validators: validators,
		proxy:      _validatorListProxy{},
	}
	set.SetMaximumPower(maximumPower)
	return set
}

func (vs *validatorSet) TotalPower() int {
	return len(vs.validators)
}

func (vs *validatorSet) SetMaximumPower(maximumPower int) {
	if maximumPower > 90 {
		maximumPower = 90
	}

	if maximumPower < 4 {
		maximumPower = 4
	}

	vs.maximumPower = maximumPower
}

func (vs *validatorSet) MaximumPower() int {
	return vs.maximumPower
}

func (vs *validatorSet) AdjustPower(height int64) error {
	/// empty slice
	vs.setLeavers = vs.setLeavers[:0]

	dif := vs.TotalPower() - vs.maximumPower
	if dif <= 0 {
		return nil
	}

	limit := int(math.Floor(float64(vs.maximumPower*1/3))) - 1
	if dif > limit {
		dif = limit
	}

	/// copy of validator set in round m
	var vals1 []acm.Validator
	var vals2 []*tm_types.Validator

	vals1 = make([]acm.Validator, len(vs.validators))
	copy(vals1, vs.validators)
	sort.SliceStable(vals1, func(i, j int) bool {
		return bytes.Compare(vals1[i].Address().Bytes(), vals1[j].Address().Bytes()) < 0
	})

	for {
		height--

		if height > 0 {
			result, err := vs.proxy.Validators(height)
			if err != nil {
				return err
			}

			/// copy of validator set in round n (n<m)
			vals2 = result.Validators
			sort.SliceStable(vals2, func(i, j int) bool {
				return bytes.Compare(vals2[i].Address.Bytes(), vals2[j].Address.Bytes()) < 0
			})
		} else {
			/// genesis validators
			vals2 = vals2[1:len(vals2)]
		}

		r := make([]int, 0)
		var i, j int = 0, 0
		for i < len(vals1) && j < len(vals2) {
			val1 := vals1[i]
			val2 := vals2[j]

			cmp := bytes.Compare(val1.Address().Bytes(), val2.Address.Bytes())
			if cmp == 0 {
				i++
				j++
			} else if cmp < 0 {
				r = append(r, i)
				i++
			} else {
				j++
			}
		}

		/// if at the end of slice_a there are some elements bigger than last element in slice_b
		for z := i; z < len(vals1); z++ {
			r = append(r, i)
		}

		// println(fmt.Sprintf("%v", vals1))
		// println(fmt.Sprintf("%v", vals2))
		// println(fmt.Sprintf("%v", r))

		var n int
		for _, m := range r {
			vals1 = append(vals1[:m-n], vals1[m-n+1:]...)
			n++

			/// Not removing more than requested
			if len(vals1) == dif {
				break
			}
		}

		if len(vals1) == dif {
			break
		}
	}

	// println(fmt.Sprintf("%v", vals1))
	for _, v1 := range vals1 {
		for i, v2 := range vs.validators {
			if v1.Address() == v2.Address() {
				vs.validators = append(vs.validators[:i], vs.validators[i+1:]...)
				vs.setLeavers = append(vs.setLeavers, v2)
				break
			}
		}
	}

	return nil
}

func (vs *validatorSet) Validators() []acm.Validator {
	return vs.validators
}

func (vs *validatorSet) SetLeavers() []acm.Validator {
	return vs.setLeavers
}

func (vs *validatorSet) JoinToTheSet(validator acm.Validator) error {
	if true == vs.IsValidatorInSet(validator.Address()) {
		return fmt.Errorf("This validator currently is in the set: %v", validator.Address())
	}

	/// Welcome to the party!
	vs.validators = append(vs.validators, validator)
	return nil
}

func (vs *validatorSet) LeaveFromTheSet(validator acm.Validator) error {
	if false == vs.IsValidatorInSet(validator.Address()) {
		return fmt.Errorf("This validator currently is not in the set: %v", validator.Address())
	}

	for i, val := range vs.validators {
		if val.Address() == validator.Address() {
			vs.validators = append(vs.validators[:i], vs.validators[i+1:]...)
			break
		}
	}

	return nil
}

func (vs *validatorSet) IsValidatorInSet(address acm.Address) bool {
	for _, v := range vs.validators {
		if v.Address() == address {
			return true
		}
	}

	return false
}

func (vs *validatorSet) JSONBytes() ([]byte, error) {
	s := make([]string, len(vs.validators))
	for i, v := range vs.validators {
		b, err := v.Bytes()

		if err != nil {
			return nil, err
		}

		s[i] = string(b)
	}

	validatorsJSON, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	u := map[string]string{}
	u["validators"] = string(validatorsJSON)
	u["maximumPower"] = strconv.FormatInt(int64(vs.maximumPower), 10)

	return json.Marshal(u)
}

func ValidatorSetFromJSON(bytes []byte) (*validatorSet, error) {
	var u map[string]string
	err := json.Unmarshal(bytes, &u)
	if err != nil {
		return nil, err
	}

	var v []string
	err = json.Unmarshal([]byte(u["validators"]), &v)
	if err != nil {
		return nil, err
	}

	validatrs := make([]acm.Validator, len(v))
	for i, s := range v {
		validatrs[i] = acm.LoadValidator([]byte(s))
	}

	maximumPower, err := strconv.ParseInt(u["maximumPower"], 10, 32)
	if err != nil {
		return nil, nil
	}

	vs := newValidatorSet(int(maximumPower), validatrs)

	return vs, nil
}
