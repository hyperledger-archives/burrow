package blockchain

import (
	"github.com/hyperledger/burrow/crypto"
)

type ValidatorsWindow struct {
	Buckets []Validators
	Total   Validators
	head    int
}

// Provides a sliding window over the last size buckets of validator power changes
func NewValidatorsWindow(size int) ValidatorsWindow {
	if size < 1 {
		size = 1
	}
	vw := ValidatorsWindow{
		Buckets: make([]Validators, size),
		Total:   NewValidators(),
	}
	vw.Buckets[vw.head] = NewValidators()
	return vw
}

// Updates the current head bucket (accumulator)
func (vw *ValidatorsWindow) AlterPower(publicKey crypto.PublicKey, power uint64) error {
	return vw.Buckets[vw.head].AlterPower(publicKey, power)
}

func (vw *ValidatorsWindow) CommitInto(validatorsToUpdate *Validators) error {
	var err error
	if vw.Buckets[vw.head].Iterate(func(publicKey crypto.PublicKey, power uint64) (stop bool) {
		// Update the sink validators
		err = validatorsToUpdate.AlterPower(publicKey, power)
		if err != nil {
			return true
		}
		// Add to total power
		err = vw.Total.AddPower(publicKey, power)
		if err != nil {
			return true
		}
		return false
	}) {
		// If iteration stopped there was an error
		return err
	}
	// move the ring buffer on
	vw.head = (vw.head + 1) % len(vw.Buckets)
	// Subtract the tail bucket (if any) from the total
	if vw.Buckets[vw.head].Iterate(func(publicKey crypto.PublicKey, power uint64) (stop bool) {
		err = vw.Total.SubtractPower(publicKey, power)
		if err != nil {
			return false
		}
		return true
	}) {
		return err
	}
	// Clear new head bucket (and possibly previous tail)
	vw.Buckets[vw.head] = NewValidators()
	return nil
}
