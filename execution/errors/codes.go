package errors

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Codes struct {
	None                   *Coding
	Generic                *Coding
	UnknownAddress         *Coding
	InsufficientBalance    *Coding
	InvalidJumpDest        *Coding
	InsufficientGas        *Coding
	MemoryOutOfBounds      *Coding
	CodeOutOfBounds        *Coding
	InputOutOfBounds       *Coding
	ReturnDataOutOfBounds  *Coding
	CallStackOverflow      *Coding
	CallStackUnderflow     *Coding
	DataStackOverflow      *Coding
	DataStackUnderflow     *Coding
	InvalidContract        *Coding
	NativeContractCodeCopy *Coding
	ExecutionAborted       *Coding
	ExecutionReverted      *Coding
	PermissionDenied       *Coding
	NativeFunction         *Coding
	EventPublish           *Coding
	InvalidString          *Coding
	EventMapping           *Coding
	InvalidAddress         *Coding
	DuplicateAddress       *Coding
	InsufficientFunds      *Coding
	Overpayment            *Coding
	ZeroPayment            *Coding
	InvalidSequence        *Coding
	ReservedAddress        *Coding
	IllegalWrite           *Coding
	IntegerOverflow        *Coding
	InvalidProposal        *Coding
	ExpiredProposal        *Coding
	ProposalExecuted       *Coding
	NoInputPermission      *Coding
	InvalidBlockNumber     *Coding
	BlockNumberOutOfRange  *Coding
	AlreadyVoted           *Coding
	UnresolvedSymbols      *Coding
	InvalidContractCode    *Coding
	NonExistentAccount     *Coding

	// For lookup
	codings []*Coding
}

func (es *Codes) init() error {
	rv := reflect.ValueOf(es).Elem()
	rt := rv.Type()
	es.codings = make([]*Coding, 0, rv.NumField())
	for i := 0; i < rv.NumField(); i++ {
		ty := rt.Field(i)
		// If field is exported
		if ty.PkgPath == "" {
			coding := rv.Field(i).Interface().(*Coding)
			if coding.Description == "" {
				return fmt.Errorf("error code '%s' has no description", ty.Name)
			}
			coding.Number = uint32(i)
			coding.Name = ty.Name
			es.codings = append(es.codings, coding)
		}
	}
	return nil
}

func (es *Codes) JSON() string {
	bs, err := json.MarshalIndent(es, "", "\t")
	if err != nil {
		panic(fmt.Errorf("could not create JSON errors object: %v", err))
	}
	return string(bs)
}

func (es *Codes) Get(number uint32) *Coding {
	if int(number) > len(es.codings)-1 {
		return nil
	}
	return es.codings[number]
}
