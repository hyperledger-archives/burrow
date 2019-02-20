package abi

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/hyperledger/burrow/deploy/compile"
	"github.com/hyperledger/burrow/execution/errors"
	log "github.com/sirupsen/logrus"
)

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

var RevertAbi *AbiSpec

func ReadAbiFormulateCallFile(abiLocation, binPath, funcName string, args []string) ([]byte, bool, error) {
	abiSpecBytes, err := readAbi(binPath, abiLocation)
	if err != nil {
		return []byte{}, false, err
	}
	log.WithField("=>", string(abiSpecBytes)).Debug("ABI Specification (Formulate)")
	log.WithFields(log.Fields{
		"function":  funcName,
		"arguments": fmt.Sprintf("%v", args),
	}).Debug("Packing Call via ABI")

	return Packer(abiSpecBytes, funcName, args...)
}

func ReadAbiFormulateCall(abiSpecBytes []byte, funcName string, args []string) ([]byte, bool, error) {
	log.WithField("=>", string(abiSpecBytes)).Debug("ABI Specification (Formulate)")
	log.WithFields(log.Fields{
		"function":  funcName,
		"arguments": fmt.Sprintf("%v", args),
	}).Debug("Packing Call via ABI")

	return Packer(string(abiSpecBytes), funcName, args...)
}

func ReadAndDecodeContractReturn(abiLocation, binPath, funcName string, resultRaw []byte) ([]*Variable, error) {
	abiSpecBytes, err := readAbi(binPath, abiLocation)
	if err != nil {
		return nil, err
	}
	log.WithField("=>", abiSpecBytes).Debug("ABI Specification (Decode)")

	// Unpack the result
	return Unpacker(abiSpecBytes, funcName, resultRaw)
}

//Convenience Packing Functions
func Packer(abiData, funcName string, args ...string) ([]byte, bool, error) {
	abiSpec, err := ReadAbiSpec([]byte(abiData))
	if err != nil {
		log.WithFields(log.Fields{
			"abi":   abiData,
			"error": err.Error(),
		}).Error("Failed to decode abi spec")
		return nil, false, err
	}

	iArgs := make([]interface{}, len(args))
	for i, s := range args {
		iArgs[i] = interface{}(s)
	}
	packedBytes, constant, err := abiSpec.Pack(funcName, iArgs...)
	if err != nil {
		log.WithFields(log.Fields{
			"abi":   abiData,
			"error": err.Error(),
		}).Error("Failed to encode abi spec")
		return nil, false, err
	}

	return packedBytes, constant, nil
}

func Unpacker(abiData, name string, data []byte) ([]*Variable, error) {
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
	for i, _ := range vals {
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

// FindEventSpec helps when you have an event but you do not know what contract
// it came from. Events can be emitted from contracts called from other contracts,
// so you are not likely to know what abi to use. Therefore, this will go through
// all the files in the directory and see if a matching event spec can be found.
func FindEventSpec(abiDir string, eventID EventID) (evAbi *EventSpec, err error) {
	err = filepath.Walk(abiDir, func(path string, fi os.FileInfo, err error) error {
		ext := filepath.Ext(path)
		if fi.IsDir() || !(ext == ".bin" || ext == ".abi") {
			return nil
		}
		if err == nil {
			abiSpc, err := ReadAbiSpecFile(path)
			if err != nil {
				return errors.Wrap(err, "Error parsing abi file "+path)
			}

			a, ok := abiSpc.EventsById[eventID]
			if ok {
				evAbi = &a
				return io.EOF
			}
		}
		return nil
	})

	if err == io.EOF {
		err = nil
	}

	return
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
