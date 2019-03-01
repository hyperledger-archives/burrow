package abi

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hyperledger/burrow/deploy/compile"
	"github.com/hyperledger/burrow/execution/errors"
	log "github.com/sirupsen/logrus"
)

// Variable exist to unpack return values into, so have both the return
// value and its name
type Variable struct {
	Name  string
	Value string
}

func init() {
	var err error
	RevertAbi, err = ReadAbiSpec([]byte(`[{"name":"Error","type":"function","outputs":[{"type":"string"}],"inputs":[{"type":"string"}]}]`))
	if err != nil {
		panic(fmt.Sprintf("internal error: failed to build revert abi: %v", err))
	}
}

// RevertAbi exists to decode reverts. Any contract function call fail using revert(), assert() or require().
// If a function exits this way, the this hardcoded ABI will be used.
var RevertAbi *AbiSpec

// EncodeFunctionCallFromFile ABI encodes a function call based on ABI in file, and the
// arguments specified as strings.
// The abiFileName specifies the name of the ABI file, and abiPath the path where it can be found.
// The fname specifies which function should called, if
// it doesn't exist exist the fallback function will be called. If fname is the empty
// string, the constructor is called. The arguments must be specified in args. The count
// must match the function being called.
// Returns the ABI encoded function call, whether the function is constant according
// to the ABI (which means it does not modified contract state)
func EncodeFunctionCallFromFile(abiFileName, abiPath, funcName string, args ...interface{}) ([]byte, *FunctionSpec, error) {
	abiSpecBytes, err := readAbi(abiPath, abiFileName)
	if err != nil {
		return []byte{}, nil, err
	}

	return EncodeFunctionCall(abiSpecBytes, funcName, args...)
}

// EncodeFunctionCall ABI encodes a function call based on ABI in string abiData
// and the arguments specified as strings.
// The fname specifies which function should called, if
// it doesn't exist exist the fallback function will be called. If fname is the empty
// string, the constructor is called. The arguments must be specified in args. The count
// must match the function being called.
// Returns the ABI encoded function call, whether the function is constant according
// to the ABI (which means it does not modified contract state)
func EncodeFunctionCall(abiData, funcName string, args ...interface{}) ([]byte, *FunctionSpec, error) {
	log.WithField("=>", abiData).Debug("ABI Specification (Formulate)")
	log.WithFields(log.Fields{
		"function":  funcName,
		"arguments": fmt.Sprintf("%v", args),
	}).Debug("Packing Call via ABI")

	abiSpec, err := ReadAbiSpec([]byte(abiData))
	if err != nil {
		log.WithFields(log.Fields{
			"abi":   abiData,
			"error": err.Error(),
		}).Error("Failed to decode abi spec")
		return nil, nil, err
	}

	packedBytes, funcSpec, err := abiSpec.Pack(funcName, args...)
	if err != nil {
		log.WithFields(log.Fields{
			"abi":   abiData,
			"error": err.Error(),
		}).Error("Failed to encode abi spec")
		return nil, nil, err
	}

	return packedBytes, funcSpec, nil
}

// DecodeFunctionReturnFromFile ABI decodes the return value from a contract function call.
func DecodeFunctionReturnFromFile(abiLocation, binPath, funcName string, resultRaw []byte) ([]*Variable, error) {
	abiSpecBytes, err := readAbi(binPath, abiLocation)
	if err != nil {
		return nil, err
	}
	log.WithField("=>", abiSpecBytes).Debug("ABI Specification (Decode)")

	// Unpack the result
	return DecodeFunctionReturn(abiSpecBytes, funcName, resultRaw)
}

func DecodeFunctionReturn(abiData, name string, data []byte) ([]*Variable, error) {
	abiSpec, err := ReadAbiSpec([]byte(abiData))
	if err != nil {
		return nil, err
	}

	var args []Argument

	if name == "" {
		args = abiSpec.Constructor.Outputs
	} else {
		if _, ok := abiSpec.Functions[name]; ok {
			args = abiSpec.Functions[name].Outputs
		} else {
			args = abiSpec.Fallback.Outputs
		}
	}

	if args == nil {
		return nil, fmt.Errorf("no such function")
	}
	vars := make([]*Variable, len(args))

	if len(args) == 0 {
		return nil, nil
	}

	vals := make([]interface{}, len(args))
	for i := range vals {
		vals[i] = new(string)
	}
	err = Unpack(args, data, vals...)
	if err != nil {
		return nil, err
	}

	for i, a := range args {
		if a.Name != "" {
			vars[i] = &Variable{Name: a.Name, Value: *(vals[i].(*string))}
		} else {
			vars[i] = &Variable{Name: fmt.Sprintf("%d", i), Value: *(vals[i].(*string))}
		}
	}

	return vars, nil
}

func readAbi(root, contract string) (string, error) {
	p := path.Join(root, stripHex(contract))
	if _, err := os.Stat(p); err != nil {
		log.WithField("abifile", p).Debug("Tried, not found")
		p = path.Join(root, stripHex(contract)+".bin")
		if _, err = os.Stat(p); err != nil {
			log.WithField("abifile", p).Debug("Tried, not found")
			return "", fmt.Errorf("Abi doesn't exist for =>\t%s", p)
		}
	}
	log.WithField("abifile", p).Debug("Found ABI")
	sol, err := compile.LoadSolidityContract(p)
	if err != nil {
		return "", err
	}
	return string(sol.Abi), nil
}

// LoadPath loads one abi file or finds all files in a directory
func LoadPath(abiFileOrDir string) (*AbiSpec, error) {
	if abiFileOrDir == "" {
		return &AbiSpec{}, fmt.Errorf("no ABI file or directory provided")
	}

	specs := make([]*AbiSpec, 0)

	for _, dir := range filepath.SplitList(abiFileOrDir) {
		err := filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("error returned while walking abiDir '%s': %v", dir, err)
			}
			ext := filepath.Ext(path)
			if fi.IsDir() || !(ext == ".bin" || ext == ".abi") {
				return nil
			}
			if err == nil {
				abiSpc, err := ReadAbiSpecFile(path)
				if err != nil {
					return errors.Wrap(err, "Error parsing abi file "+path)
				}
				specs = append(specs, abiSpc)
			}
			return nil
		})
		if err != nil {
			return &AbiSpec{}, err
		}
	}
	return MergeAbiSpec(specs), nil
}

func stripHex(s string) string {
	if len(s) > 1 {
		if s[:2] == "0x" {
			s = s[2:]
			if len(s)%2 != 0 {
				s = "0" + s
			}
			return s
		}
	}
	return s
}
