package blockchain

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/crypto"
)

type ValidatorsRing struct {
	buckets []*Validators
	// Totals for each validator across all buckets
	power *Validators
	// Current flow totals for each validator in the Head bucket
	flow *Validators
	// Index of current head bucekt
	head int64
	size int64
}

var big1 = big.NewInt(1)
var big3 = big.NewInt(3)

// Provides a sliding window over the last size buckets of validator power changes
func NewValidatorsRing(initialValidators *Validators, size int) *ValidatorsRing {
	if size < 2 {
		size = 2
	}
	vw := &ValidatorsRing{
		buckets: make([]*Validators, size),
		power:   NewValidators(),
		flow:    NewValidators(),
		size:    int64(size),
	}
	for i := 0; i < size; i++ {
		vw.buckets[i] = NewValidators()
	}

	initialValidators.Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		// Existing set
		vw.buckets[vw.index(-1)].AlterPower(id, power)
		// Current accumulator
		vw.buckets[vw.head].AlterPower(id, power)

		vw.power.AddPower(id, power.Add(power, power))
		return
	})

	return vw
}

// Updates the current head bucket (accumulator) with some safety checks
func (vw *ValidatorsRing) AlterPower(id crypto.Addressable, power *big.Int) (*big.Int, error) {
	if power.Sign() == -1 {
		return nil, fmt.Errorf("cannot set negative validator power: %v", power)
	}
	// if flow > maxflow then we cannot alter the power
	flow := vw.Flow(id, power)
	maxFlow := vw.MaxFlow()
	// Set flow to update total flow
	vw.flow.AlterPower(id, flow)
	if vw.flow.totalPower.Cmp(maxFlow) == 1 {
		// Reset flow to previous value
		vw.flow.AlterPower(id, vw.Flow(id, vw.Head().Power(id)))
		allowable := new(big.Int).Sub(maxFlow, vw.flow.totalPower)
		return nil, fmt.Errorf("cannot change validator power of %v from %v to %v because that would result in a flow "+
			"greater than or equal to 1/3 of total power for the next commit: flow induced by change: %v, "+
			"current total flow: %v/%v (cumulative/max), remaining allowable flow: %v",
			id.Address(), vw.Prev().Power(id), power, flow, vw.flow.totalPower, maxFlow, allowable)
	}
	// Add to total power
	vw.Head().AlterPower(id, power)
	return flow, nil
}

// Returns the flow that would be induced by a validator change by comparing the current accumulator with the previous
// bucket
func (vw *ValidatorsRing) Flow(id crypto.Addressable, power *big.Int) *big.Int {
	flow := new(big.Int)
	prevPower := vw.Prev().Power(id)
	return flow.Abs(flow.Sub(power, prevPower))
}

// To ensure that in the maximum valildator shift at least one unit
// of validator power in the intersection of last block validators and this block validators must have at least one
// non-byzantine validator who can tell you if you've been lied to about the validator set
// So need at most ceiling((Total Power)/3) - 1, in integer division we have ceiling(X*p/q) = (p(X+1)-1)/q
// For p = 1 just X/q
// So we want (Total Power)/3 - 1
func (vw *ValidatorsRing) MaxFlow() *big.Int {
	max := vw.Prev().TotalPower()
	return max.Sub(max.Div(max, big3), big1)
}

// Advance the current head bucket to the next bucket and returns the change in total power between the previous bucket
// and the current head, and the total flow which is the sum of absolute values of all changes each validator's power
// after rotation the next head is a copy of the current head
func (vw *ValidatorsRing) Rotate() (totalPowerChange *big.Int, totalFlow *big.Int) {
	// Subtract the previous bucket total power so we can add on the current buckets power after this
	totalPowerChange = new(big.Int).Sub(vw.Head().totalPower, vw.Prev().totalPower)
	// Capture flow before we wipe it
	totalFlow = vw.flow.totalPower
	// Subtract the tail bucket (if any) from the total
	vw.power.Subtract(vw.Next())
	// Copy head bucket
	headCopy := vw.Head().Copy()
	//	add it to total
	vw.power.Add(headCopy)
	// move the ring buffer on
	vw.head = vw.index(1)
	// Overwrite new head bucket (previous tail) with previous bucket copy updated with current head
	vw.buckets[vw.head] = headCopy
	// New flow accumulator1
	vw.flow = NewValidators()
	// Advance the ring
	return totalPowerChange, totalFlow
}

func (vw *ValidatorsRing) Prev() *Validators {
	return vw.buckets[vw.index(-1)]
}

func (vw *ValidatorsRing) Head() *Validators {
	return vw.buckets[vw.head]
}

func (vw *ValidatorsRing) Next() *Validators {
	return vw.buckets[vw.index(1)]
}

func (vw *ValidatorsRing) index(i int64) int64 {
	return (vw.size + vw.head + i) % vw.size
}

func (vw *ValidatorsRing) Size() int64 {
	return vw.size
}

// Returns buckets in order head, previous, ...
func (vw *ValidatorsRing) OrderedBuckets() []*Validators {
	buckets := make([]*Validators, len(vw.buckets))
	for i := int64(0); i < vw.size; i++ {
		buckets[i] = vw.buckets[vw.index(-i)]
	}
	return buckets
}

func (vw *ValidatorsRing) String() string {
	return fmt.Sprintf("ValidatorsWindow{Total: %v; Buckets: Head->%v<-Tail}", vw.power, vw.OrderedBuckets())
}

func (vw *ValidatorsRing) Equal(vwOther *ValidatorsRing) bool {
	if vw.size != vwOther.size || vw.head != vwOther.head || len(vw.buckets) != len(vwOther.buckets) ||
		!vw.flow.Equal(vwOther.flow) || !vw.power.Equal(vwOther.power) {
		return false
	}
	for i, b := range vw.buckets {
		if !b.Equal(vwOther.buckets[i]) {
			return false
		}
	}
	return true
}

type PersistedValidatorsRing struct {
	Buckets [][]PersistedValidator
	Power   []PersistedValidator
	Flow    []PersistedValidator
	Head    int64
}

func (vw *ValidatorsRing) Persistable() PersistedValidatorsRing {
	buckets := make([][]PersistedValidator, len(vw.buckets))
	for i, vs := range vw.buckets {
		buckets[i] = vs.Persistable()
	}
	return PersistedValidatorsRing{
		Buckets: buckets,
		Power:   vw.power.Persistable(),
		Flow:    vw.flow.Persistable(),
		Head:    vw.head,
	}
}

func UnpersistValidatorsRing(pvr PersistedValidatorsRing) *ValidatorsRing {
	buckets := make([]*Validators, len(pvr.Buckets))
	for i, pv := range pvr.Buckets {
		buckets[i] = UnpersistValidators(pv)
	}

	return &ValidatorsRing{
		buckets: buckets,
		head:    pvr.Head,
		power:   UnpersistValidators(pvr.Power),
		flow:    UnpersistValidators(pvr.Flow),
		size:    int64(len(buckets)),
	}
}
