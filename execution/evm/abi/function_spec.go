package abi

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto/sha3"
)

// FunctionIDSize is the length of the function selector
const FunctionIDSize = 4

type FunctionSpec struct {
	FunctionID FunctionID
	Constant   bool
	Inputs     []Argument
	Outputs    []Argument
}

type FunctionID [FunctionIDSize]byte

func GetFunctionID(signature string) (id FunctionID) {
	hash := sha3.NewKeccak256()
	hash.Write([]byte(signature))
	copy(id[:], hash.Sum(nil)[:4])
	return
}

func Signature(name string, args []Argument) string {
	return name + argsToSignature(args, false)
}

func (f *FunctionSpec) String(name string) string {
	return name + argsToSignature(f.Inputs, true) +
		" returns " + argsToSignature(f.Outputs, true)
}

func (f *FunctionSpec) SetFunctionID(functionName string) {
	sig := Signature(functionName, f.Inputs)
	f.FunctionID = GetFunctionID(sig)
}

func (fs FunctionID) Bytes() []byte {
	return fs[:]
}

func argsToSignature(args []Argument, addIndexedName bool) (str string) {
	str = "("
	for i, a := range args {
		if i > 0 {
			str += ","
		}
		str += a.EVM.GetSignature()
		if addIndexedName && a.Indexed {
			str += " indexed"
		}
		if a.IsArray {
			if a.ArrayLength > 0 {
				str += fmt.Sprintf("[%d]", a.ArrayLength)
			} else {
				str += "[]"
			}
		}
		if addIndexedName && a.Name != "" {
			str += " " + a.Name
		}
	}
	str += ")"
	return
}
