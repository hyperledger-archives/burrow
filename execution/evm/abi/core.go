package abi

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/hyperledger/burrow/deploy/compile"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/logging"
)

// Variable exist to unpack return values into, so have both the return
// value and its name
type Variable struct {
	Name  string
	Value string
}

func init() {
	var err error
	RevertAbi, err = ReadSpec([]byte(`[{"name":"Error","type":"function","outputs":[{"type":"string"}],"inputs":[{"type":"string"}]}]`))
	if err != nil {
		panic(fmt.Sprintf("internal error: failed to build revert abi: %v", err))
	}
}

// RevertAbi exists to decode reverts. Any contract function call fail using revert(), assert() or require().
// If a function exits this way, the this hardcoded ABI will be used.
var RevertAbi *Spec

// EncodeFunctionCallFromFile ABI encodes a function call based on ABI in file, and the
// arguments specified as strings.
// The abiFileName specifies the name of the ABI file, and abiPath the path where it can be found.
// The fname specifies which function should called, if
// it doesn't exist exist the fallback function will be called. If fname is the empty
// string, the constructor is called. The arguments must be specified in args. The count
// must match the function being called.
// Returns the ABI encoded function call, whether the function is constant according
// to the ABI (which means it does not modified contract state)
func EncodeFunctionCallFromFile(abiFileName, abiPath, funcName string, logger *logging.Logger, args ...interface{}) ([]byte, *FunctionSpec, error) {
	abiSpecBytes, err := readAbi(abiPath, abiFileName, logger)
	if err != nil {
		return []byte{}, nil, err
	}

	return EncodeFunctionCall(abiSpecBytes, funcName, logger, args...)
}

// EncodeFunctionCall ABI encodes a function call based on ABI in string abiData
// and the arguments specified as strings.
// The fname specifies which function should called, if
// it doesn't exist exist the fallback function will be called. If fname is the empty
// string, the constructor is called. The arguments must be specified in args. The count
// must match the function being called.
// Returns the ABI encoded function call, whether the function is constant according
// to the ABI (which means it does not modified contract state)
func EncodeFunctionCall(abiData, funcName string, logger *logging.Logger, args ...interface{}) ([]byte, *FunctionSpec, error) {
	logger.TraceMsg("Packing Call via ABI",
		"spec", abiData,
		"function", funcName,
		"arguments", fmt.Sprintf("%v", args),
	)

	abiSpec, err := ReadSpec([]byte(abiData))
	if err != nil {
		logger.InfoMsg("Failed to decode abi spec",
			"abi", abiData,
			"error", err.Error(),
		)
		return nil, nil, err
	}

	packedBytes, funcSpec, err := abiSpec.Pack(funcName, args...)
	if err != nil {
		logger.InfoMsg("Failed to encode abi spec",
			"abi", abiData,
			"error", err.Error(),
		)
		return nil, nil, err
	}

	return packedBytes, funcSpec, nil
}

// DecodeFunctionReturnFromFile ABI decodes the return value from a contract function call.
func DecodeFunctionReturnFromFile(abiLocation, binPath, funcName string, resultRaw []byte, logger *logging.Logger) ([]*Variable, error) {
	abiSpecBytes, err := readAbi(binPath, abiLocation, logger)
	if err != nil {
		return nil, err
	}
	logger.TraceMsg("ABI Specification (Decode)", "spec", abiSpecBytes)

	// Unpack the result
	return DecodeFunctionReturn(abiSpecBytes, funcName, resultRaw)
}

func DecodeFunctionReturn(abiData, name string, data []byte) ([]*Variable, error) {
	abiSpec, err := ReadSpec([]byte(abiData))
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

func readAbi(root, contract string, logger *logging.Logger) (string, error) {
	p := path.Join(root, stripHex(contract))
	if _, err := os.Stat(p); err != nil {
		logger.TraceMsg("abifile not found", "tried", p)
		p = path.Join(root, stripHex(contract)+".bin")
		if _, err = os.Stat(p); err != nil {
			logger.TraceMsg("abifile not found", "tried", p)
			return "", fmt.Errorf("abi doesn't exist for =>\t%s", p)
		}
	}
	logger.TraceMsg("Found ABI file", "path", p)
	sol, err := compile.LoadSolidityContract(p)
	if err != nil {
		return "", err
	}
	return string(sol.Abi), nil
}

// LoadPath loads one abi file or finds all files in a directory
func LoadPath(abiFileOrDirs ...string) (*Spec, error) {
	if len(abiFileOrDirs) == 0 {
		return &Spec{}, fmt.Errorf("no ABI file or directory provided")
	}

	specs := make([]*Spec, 0)

	for _, dir := range abiFileOrDirs {
		err := filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("error returned while walking abiDir '%s': %v", dir, err)
			}
			ext := filepath.Ext(path)
			if fi.IsDir() || !(ext == ".bin" || ext == ".abi") {
				return nil
			}
			if err == nil {
				abiSpc, err := ReadSpecFile(path)
				if err != nil {
					return errors.Wrap(err, "Error parsing abi file "+path)
				}
				specs = append(specs, abiSpc)
			}
			return nil
		})
		if err != nil {
			return &Spec{}, err
		}
	}
	return MergeSpec(specs), nil
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
