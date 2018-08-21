package validator

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/crypto"
)

type Ring struct {
	// The validator power history stored in buckets as a ring buffer
	// The changes committed at rotation i
	delta []*Set
	// The cumulative changes at rotation i - 1
	cum []*Set
	// Totals for each validator across all buckets
	power *Set
	// Current flow totals for each validator in the Head bucket
	flow *Set
	// Index of current head bucket
	head int64
	// Number of buckets
	size int64
}

var big1 = big.NewInt(1)
var big3 = big.NewInt(3)

// Provides a sliding window over the last size buckets of validator power changes
func NewRing(initialSet Iterable, windowSize int) *Ring {
	if windowSize < 1 {
		windowSize = 1
	}
	vw := &Ring{
		delta: make([]*Set, windowSize),
		cum:   make([]*Set, windowSize),
		power: NewSet(),
		flow:  NewSet(),
		size:  int64(windowSize),
	}
	for i := 0; i < windowSize; i++ {
		vw.delta[i] = NewSet()
		// Important that this is trim set for accurate count
		vw.cum[i] = NewTrimSet()
	}
	vw.cum[0] = Copy(initialSet)

	return vw
}

// Implement Reader
// Get power at index from the delta bucket then falling through to the cumulative
func (vc *Ring) PowerAt(index int64, id crypto.Address) *big.Int {
	power := vc.Head().MaybePower(id)
	if power != nil {
		return power
	}
	return vc.Cum().Power(id)
}

func (vc *Ring) Power(id crypto.Address) *big.Int {
	return vc.PowerAt(vc.head, id)
}

// Return the resultant set at index of current cum plus delta
func (vc *Ring) Resultant(index int64) *Set {
	i := vc.index(index)
	cum := CopyTrim(vc.cum[i])
	vc.delta[i].Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		cum.AlterPower(id.PublicKey(), power)
		return
	})
	return cum
}

func (vc *Ring) TotalPower() *big.Int {
	return vc.Resultant(vc.head).totalPower
}

// Updates the current head bucket (accumulator) with some safety checks
func (vc *Ring) AlterPower(id crypto.PublicKey, power *big.Int) (*big.Int, error) {
	if power.Sign() == -1 {
		return nil, fmt.Errorf("cannot set negative validator power: %v", power)
	}
	if !power.IsInt64() {
		return nil, fmt.Errorf("for tendermint compatibility validator power must fit within an Int64 bur %v "+
			"does not", power)
	}
	// if flow > maxflow then we cannot alter the power
	flow := vc.Flow(id.Address(), power)
	maxFlow := vc.MaxFlow()
	// Set flow for this id to update flow.totalPower (total flow) for comparison below, keep track of flow for each id
	// so that we only count flow once for each id
	vc.flow.ChangePower(id, flow)
	// The totalPower of the Flow Set is the absolute value of all power changes made so far
	if vc.flow.totalPower.Cmp(maxFlow) == 1 {
		// Reset flow to previous value to undo update above
		prevFlow := vc.Flow(id.Address(), vc.Head().Power(id.Address()))
		vc.flow.ChangePower(id, prevFlow)
		allowable := new(big.Int).Sub(maxFlow, vc.flow.totalPower)
		return nil, fmt.Errorf("cannot change validator power of %v from %v to %v because that would result in a flow "+
			"greater than or equal to 1/3 of total power for the next commit: flow induced by change: %v, "+
			"current total flow: %v/%v (cumulative/max), remaining allowable flow: %v",
			id.Address(), vc.Cum().Power(id.Address()), power, flow, vc.flow.totalPower, maxFlow, allowable)
	}
	// Add to total power
	vc.Head().ChangePower(id, power)
	return flow, nil
}

// Returns the flow that would be induced by a validator change by comparing the head accumulater with the current set
func (vc *Ring) Flow(id crypto.Address, power *big.Int) *big.Int {
	flow := new(big.Int)
	return flow.Abs(flow.Sub(power, vc.Cum().Power(id)))
}

// To ensure that in the maximum valildator shift at least one unit
// of validator power in the intersection of last block validators and this block validators must have at least one
// non-byzantine validator who can tell you if you've been lied to about the validator set
// So need at most ceiling((Total Power)/3) - 1, in integer division we have ceiling(X*p/q) = (p(X+1)-1)/q
// For p = 1 just X/q
// So we want (Total Power)/3 - 1
func (vc *Ring) MaxFlow() *big.Int {
	max := vc.Cum().TotalPower()
	return max.Sub(max.Div(max, big3), big1)
}

// Advance the current head bucket to the next bucket and returns the change in total power between the previous bucket
// and the current head, and the total flow which is the sum of absolute values of all changes each validator's power
// after rotation the next head is a copy of the current head
func (vc *Ring) Rotate() (totalPowerChange *big.Int, totalFlow *big.Int, err error) {
	// Subtract the tail bucket (if any) from the total
	err = Subtract(vc.power, vc.Next())
	if err != nil {
		return
	}
	// Add head delta to total power
	err = Add(vc.power, vc.Head())
	if err != nil {
		return
	}
	// Copy current cumulative bucket
	cum := CopyTrim(vc.Cum())
	// Copy delta into what will be the next cumulative bucket
	err = Alter(cum, vc.Head())
	if err != nil {
		return
	}
	// Advance the ring buffer
	vc.head = vc.index(1)
	// Overwrite new head bucket (previous tail) with a fresh delta accumulator
	vc.delta[vc.head] = NewSet()
	// Set the next cum
	vc.cum[vc.head] = cum
	// Capture flow before we wipe it
	totalFlow = vc.flow.totalPower
	// New flow accumulator
	vc.flow = NewSet()
	// Subtract the previous bucket total power so we can add on the current buckets power after this
	totalPowerChange = new(big.Int).Sub(vc.Cum().TotalPower(), vc.cum[vc.index(-1)].TotalPower())
	return
}

func (vc *Ring) CurrentSet() *Set {
	return vc.cum[vc.head]
}

func (vc *Ring) PreviousSet() *Set {
	return vc.cum[vc.index(-1)]
}

func (vc *Ring) Cum() *Set {
	return vc.cum[vc.head]
}

// Get the current accumulator bucket
func (vc *Ring) Head() *Set {
	return vc.delta[vc.head]
}

func (vc *Ring) Next() *Set {
	return vc.delta[vc.index(1)]
}

func (vc *Ring) index(i int64) int64 {
	idx := (vc.size + vc.head + i) % vc.size
	return idx
}

// Get the number of buckets in the ring (use Current().Count() to get the current number of validators)
func (vc *Ring) Size() int64 {
	return vc.size
}

// Returns buckets in order head, previous, ...
func (vc *Ring) OrderedBuckets() (delta, cum []*Set) {
	delta = make([]*Set, len(vc.delta))
	cum = make([]*Set, len(vc.cum))
	for i := int64(0); i < vc.size; i++ {
		index := vc.index(-i)
		delta[i] = vc.delta[index]
		cum[i] = vc.cum[index]
	}
	return
}

func (vc *Ring) String() string {
	delta, _ := vc.OrderedBuckets()
	return fmt.Sprintf("ValidatorsWindow{Total: %v; Delta: Head->%v<-Tail}", vc.power, delta)
}

func (vc *Ring) Equal(vwOther *Ring) bool {
	if vc.size != vwOther.size || vc.head != vwOther.head || len(vc.delta) != len(vwOther.delta) ||
		!vc.flow.Equal(vwOther.flow) || !vc.power.Equal(vwOther.power) {
		return false
	}
	for i := 0; i < len(vc.delta); i++ {
		if !vc.delta[i].Equal(vwOther.delta[i]) || !vc.cum[i].Equal(vwOther.cum[i]) {
			return false
		}
	}
	return true
}

type PersistedRing struct {
	Delta [][]*Validator
	Cum   [][]*Validator
	Power []*Validator
	Flow  []*Validator
	Head  int64
}

func (vc *Ring) Persistable() PersistedRing {
	delta := make([][]*Validator, len(vc.delta))
	cum := make([][]*Validator, len(vc.cum))
	for i := 0; i < len(delta); i++ {
		delta[i] = vc.delta[i].Validators()
		cum[i] = vc.cum[i].Validators()

	}
	return PersistedRing{
		Delta: delta,
		Cum:   cum,
		Power: vc.power.Validators(),
		Flow:  vc.flow.Validators(),
		Head:  vc.head,
	}
}

func UnpersistRing(pc PersistedRing) *Ring {
	delta := make([]*Set, len(pc.Delta))
	cum := make([]*Set, len(pc.Cum))
	for i := 0; i < len(delta); i++ {
		delta[i] = UnpersistSet(pc.Delta[i])
		cum[i] = UnpersistSet(pc.Cum[i])
	}
	return &Ring{
		delta: delta,
		cum:   cum,
		head:  pc.Head,
		power: UnpersistSet(pc.Power),
		flow:  UnpersistSet(pc.Flow),
		size:  int64(len(delta)),
	}
}
