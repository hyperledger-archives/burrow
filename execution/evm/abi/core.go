package abi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/hyperledger/burrow/deploy/compile"
	log "github.com/sirupsen/logrus"
)

type Variable struct {
	Name  string
	Value string
}

func ReadAbiFormulateCallFile(abiLocation, binPath, funcName string, args []string) ([]byte, error) {
	abiSpecBytes, err := readAbi(binPath, abiLocation)
	if err != nil {
		return []byte{}, err
	}
	log.WithField("=>", string(abiSpecBytes)).Debug("ABI Specification (Formulate)")
	log.WithFields(log.Fields{
		"function":  funcName,
		"arguments": fmt.Sprintf("%v", args),
	}).Debug("Packing Call via ABI")

	return Packer(abiSpecBytes, funcName, args...)
}

func ReadAbiFormulateCall(abiSpecBytes []byte, funcName string, args []string) ([]byte, error) {
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
func Packer(abiData, funcName string, args ...string) ([]byte, error) {
	abiSpec, err := ReadAbiSpec([]byte(abiData))
	if err != nil {
		log.WithFields(log.Fields{
			"abi":   abiData,
			"error": err.Error(),
		}).Error("Failed to decode abi spec")
		return nil, err
	}

	iArgs := make([]interface{}, len(args))
	for i, s := range args {
		iArgs[i] = interface{}(s)
	}
	packedBytes, err := abiSpec.Pack(funcName, iArgs...)
	if err != nil {
		log.WithFields(log.Fields{
			"abi":   abiData,
			"error": err.Error(),
		}).Error("Failed to encode abi spec")
		return nil, err
	}

	return packedBytes, nil
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
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return "", err
	}
	sol := compile.SolidityOutputContract{}
	err = json.Unmarshal(b, &sol)
	if err != nil {
		return "", err
	}

	return string(sol.Abi), nil
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
