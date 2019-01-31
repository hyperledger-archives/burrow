package sqlsol

import (
	"os"
	"path/filepath"

	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/pkg/errors"
)

// AbiLoader loads abi files and parses them
func AbiLoader(abiDir, abiFile string) (*abi.AbiSpec, error) {

	var abiSpec *abi.AbiSpec
	var err error

	if abiDir == "" && abiFile == "" {
		return &abi.AbiSpec{}, errors.New("One of AbiDir or AbiFile must be provided")
	}

	if abiDir != "" && abiFile != "" {
		return &abi.AbiSpec{}, errors.New("AbiDir or AbiFile must be provided, but not both")
	}

	if abiDir != "" {
		specs := make([]*abi.AbiSpec, 0)

		err := filepath.Walk(abiDir, func(path string, fi os.FileInfo, err error) error {
			ext := filepath.Ext(path)
			if fi.IsDir() || !(ext == ".bin" || ext == ".abi") {
				return nil
			}
			if err == nil {
				abiSpc, err := abi.ReadAbiSpecFile(path)
				if err != nil {
					return errors.Wrap(err, "Error parsing abi file "+path)
				}
				specs = append(specs, abiSpc)
			}
			return nil
		})
		if err != nil {
			return &abi.AbiSpec{}, err
		}
		abiSpec = abi.MergeAbiSpec(specs)
	} else {
		abiSpec, err = abi.ReadAbiSpecFile(abiFile)
		if err != nil {
			return &abi.AbiSpec{}, errors.Wrap(err, "Error parsing abi file")
		}
	}

	return abiSpec, nil
}
