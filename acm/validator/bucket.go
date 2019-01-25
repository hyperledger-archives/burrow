package validator

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/crypto"
)

type Bucket struct {
	// Delta tracks the changes to validator power made since the previous rotation
	Delta *Set
	// Cumulative tracks the value for all validator powers at the point of the last rotation (the sum of all the deltas over all rotations) - these are the history of the complete validator sets at each rotation
	Cum *Set
	// Flow tracks the absolute value of all flows (difference between previous cum bucket and current delta) towards and away from each validator (tracking each validator separately to avoid double counting flows made against the same validator
	Flow *Set
}

func NewBucket(initialSet ...Iterable) *Bucket {
	return &Bucket{
		Cum:   CopyTrim(initialSet...),
		Delta: NewSet(),
		Flow:  NewSet(),
	}
}

// Implement Reader
func (vc *Bucket) Power(id crypto.Address) (*big.Int, error) {
	return vc.Cum.Power(id)
}

// Updates the current head bucket (accumulator) whilst
func (vc *Bucket) AlterPower(id crypto.PublicKey, power *big.Int) (*big.Int, error) {
	err := checkPower(power)
	if err != nil {
		return nil, err
	}
	// The max flow we are permitted to allow across all validators
	maxFlow := vc.Cum.MaxFlow()
	// The remaining flow we have to play with
	allowableFlow := new(big.Int).Sub(maxFlow, vc.Flow.totalPower)
	// The new absolute flow caused by this AlterPower
	absFlow := new(big.Int).Abs(vc.Cum.Flow(id, power))
	// If we call vc.flow.ChangePower(id, absFlow) (below) will we induce a change in flow greater than the allowable
	// flow we have left to spend?
	if vc.Flow.Flow(id, absFlow).Cmp(allowableFlow) == 1 {
		return nil, fmt.Errorf("cannot change validator power of %v from %v to %v because that would result in a flow "+
			"greater than or equal to 1/3 of total power for the next commit: flow induced by change: %v, "+
			"current total flow: %v/%v (cumulative/max), remaining allowable flow: %v",
			id.GetAddress(), vc.Cum.GetPower(id.GetAddress()), power, absFlow, vc.Flow.totalPower, maxFlow, allowableFlow)
	}
	// Set flow for this id to update flow.totalPower (total flow) for comparison below, keep track of flow for each id
	// so that we only count flow once for each id
	vc.Flow.ChangePower(id, absFlow)
	// Add to total power
	vc.Delta.ChangePower(id, power)
	return absFlow, nil
}

func (vc *Bucket) SetPower(id crypto.PublicKey, power *big.Int) error {
	err := checkPower(power)
	if err != nil {
		return err
	}
	// The new absolute flow caused by this AlterPower
	absFlow := new(big.Int).Abs(vc.Cum.Flow(id, power))
	// Set flow for this id to update flow.totalPower (total flow) for comparison below, keep track of flow for each id
	// so that we only count flow once for each id
	vc.Flow.ChangePower(id, absFlow)
	// Add to total power
	vc.Delta.ChangePower(id, power)
	return nil
}

func (vc *Bucket) CurrentSet() *Set {
	return vc.Cum
}

func (vc *Bucket) NextSet() *Set {
	return Copy(vc.Cum, vc.Delta)
}

func (vc *Bucket) String() string {
	return fmt.Sprintf("Bucket{Cum: %v; Delta: %v}", vc.Cum, vc.Delta)
}

func (vc *Bucket) Equal(vwOther *Bucket) error {
	err := vc.Delta.Equal(vwOther.Delta)
	if err != nil {
		return fmt.Errorf("bucket delta != other bucket delta: %v", err)
	}
	err = vc.Cum.Equal(vwOther.Cum)
	if err != nil {
		return fmt.Errorf("bucket cum != other bucket cum: %v", err)
	}
	return nil
}

func checkPower(power *big.Int) error {
	if power.Sign() == -1 {
		return fmt.Errorf("cannot set negative validator power: %v", power)
	}
	if !power.IsInt64() {
		return fmt.Errorf("for tendermint compatibility validator power must fit within an int but %v "+
			"does not", power)
	}
	return nil
}
