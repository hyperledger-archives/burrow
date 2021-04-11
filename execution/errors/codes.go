package errors

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type codes struct {
	None                   *Code
	Generic                *Code
	UnknownAddress         *Code
	InsufficientBalance    *Code
	InvalidJumpDest        *Code
	InsufficientGas        *Code
	MemoryOutOfBounds      *Code
	CodeOutOfBounds        *Code
	InputOutOfBounds       *Code
	ReturnDataOutOfBounds  *Code
	CallStackOverflow      *Code
	CallStackUnderflow     *Code
	DataStackOverflow      *Code
	DataStackUnderflow     *Code
	InvalidContract        *Code
	NativeContractCodeCopy *Code
	ExecutionAborted       *Code
	ExecutionReverted      *Code
	PermissionDenied       *Code
	NativeFunction         *Code
	EventPublish           *Code
	InvalidString          *Code
	EventMapping           *Code
	InvalidAddress         *Code
	DuplicateAddress       *Code
	InsufficientFunds      *Code
	Overpayment            *Code
	ZeroPayment            *Code
	InvalidSequence        *Code
	ReservedAddress        *Code
	IllegalWrite           *Code
	IntegerOverflow        *Code
	InvalidProposal        *Code
	ExpiredProposal        *Code
	ProposalExecuted       *Code
	NoInputPermission      *Code
	InvalidBlockNumber     *Code
	BlockNumberOutOfRange  *Code
	AlreadyVoted           *Code
	UnresolvedSymbols      *Code
	InvalidContractCode    *Code
	NonExistentAccount     *Code
	NotCallable            *Code

	// For lookup
	codes []*Code
}

func (es *codes) init() error {
	rv := reflect.ValueOf(es).Elem()
	rt := rv.Type()
	es.codes = make([]*Code, 0, rv.NumField())
	for i := 0; i < rv.NumField(); i++ {
		ty := rt.Field(i)
		// If field is exported
		if ty.PkgPath == "" {
			coding := rv.Field(i).Interface().(*Code)
			if coding.Description == "" {
				return fmt.Errorf("error code '%s' has no description", ty.Name)
			}
			coding.Number = uint32(i)
			coding.Name = ty.Name
			es.codes = append(es.codes, coding)
		}
	}
	return nil
}

func (es *codes) JSON() string {
	bs, err := json.MarshalIndent(es, "", "\t")
	if err != nil {
		panic(fmt.Errorf("could not create JSON errors object: %v", err))
	}
	return string(bs)
}

func (es *codes) Get(number uint32) *Code {
	if int(number) > len(es.codes)-1 {
		return nil
	}
	return es.codes[number]
}

func (es *codes) String() string {
	sb := new(strings.Builder)
	for _, c := range es.codes {
		sb.WriteString(strconv.FormatUint(uint64(c.Number), 10))
		sb.WriteString(": ")
		sb.WriteString(c.Name)
		sb.WriteString(" - ")
		sb.WriteString(c.Description)
		sb.WriteRune('\n')
	}
	return sb.String()
}
