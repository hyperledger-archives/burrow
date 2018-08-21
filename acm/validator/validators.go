package validator

import (
	"math/big"

	"sync"

	"github.com/hyperledger/burrow/crypto"
)

type Writer interface {
	AlterPower(id crypto.PublicKey, power *big.Int) (flow *big.Int, err error)
}

type Reader interface {
	Power(id crypto.Address) *big.Int
}

type Iterable interface {
	Iterate(func(id crypto.Addressable, power *big.Int) (stop bool)) (stopped bool)
}

type IterableReader interface {
	Reader
	Iterable
}

type ReaderWriter interface {
	Reader
	Writer
}

type IterableReaderWriter interface {
	ReaderWriter
	Iterable
}

type WriterFunc func(id crypto.PublicKey, power *big.Int) (flow *big.Int, err error)

func SyncWriter(locker sync.Locker, writerFunc WriterFunc) WriterFunc {
	return WriterFunc(func(id crypto.PublicKey, power *big.Int) (flow *big.Int, err error) {
		locker.Lock()
		defer locker.Unlock()
		return writerFunc(id, power)
	})
}

func (wf WriterFunc) AlterPower(id crypto.PublicKey, power *big.Int) (flow *big.Int, err error) {
	return wf(id, power)
}

func AddPower(vs ReaderWriter, id crypto.PublicKey, power *big.Int) error {
	// Current power + power
	_, err := vs.AlterPower(id, new(big.Int).Add(vs.Power(id.Address()), power))
	return err
}

func SubtractPower(vs ReaderWriter, id crypto.PublicKey, power *big.Int) error {
	_, err := vs.AlterPower(id, new(big.Int).Sub(vs.Power(id.Address()), power))
	return err
}

func Alter(vs Writer, vsOther Iterable) (err error) {
	vsOther.Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		_, err = vs.AlterPower(id.PublicKey(), power)
		if err != nil {
			return true
		}
		return
	})
	return
}

// Adds vsOther to vs
func Add(vs ReaderWriter, vsOther Iterable) (err error) {
	vsOther.Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		err = AddPower(vs, id.PublicKey(), power)
		if err != nil {
			return true
		}
		return
	})
	return
}

// Subtracts vsOther from vs
func Subtract(vs ReaderWriter, vsOther Iterable) (err error) {
	vsOther.Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		err = SubtractPower(vs, id.PublicKey(), power)
		if err != nil {
			return true
		}
		return
	})
	return
}

func Copy(vs Iterable) *Set {
	vsCopy := NewSet()
	vs.Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		vsCopy.ChangePower(id.PublicKey(), power)
		return
	})
	return vsCopy
}

func CopyTrim(vs Iterable) *Set {
	s := Copy(vs)
	s.trim = true
	return s
}
