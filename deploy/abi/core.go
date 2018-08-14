package abi

import (
	"fmt"

	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/util"
	log "github.com/sirupsen/logrus"
)

func ReadAbiFormulateCallFile(abiLocation string, funcName string, args []string, do *def.Packages) ([]byte, error) {
	abiSpecBytes, err := util.ReadAbi(do.BinPath, abiLocation)
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

func ReadAbiFormulateCall(abiSpecBytes []byte, funcName string, args []string, do *def.Packages) ([]byte, error) {
	log.WithField("=>", string(abiSpecBytes)).Debug("ABI Specification (Formulate)")
	log.WithFields(log.Fields{
		"function":  funcName,
		"arguments": fmt.Sprintf("%v", args),
	}).Debug("Packing Call via ABI")

	return Packer(string(abiSpecBytes), funcName, args...)
}

func ReadAndDecodeContractReturn(abiLocation, funcName string, resultRaw []byte, do *def.Packages) ([]*def.Variable, error) {
	abiSpecBytes, err := util.ReadAbi(do.BinPath, abiLocation)
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

func Unpacker(abiData, name string, data []byte) ([]*def.Variable, error) {
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
	vars := make([]*def.Variable, len(args))

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
			vars[i] = &def.Variable{Name: a.Name, Value: *(vals[i].(*string))}
		} else {
			vars[i] = &def.Variable{Name: fmt.Sprintf("%d", i), Value: *(vals[i].(*string))}
		}
	}

	return vars, nil
}
