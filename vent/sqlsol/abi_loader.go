package sqlsol

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/pkg/errors"
)

// AbiLoader loads abi files and parses them
func AbiLoader(abiFileOrDir string) (*abi.AbiSpec, error) {
	if abiFileOrDir == "" {
		return &abi.AbiSpec{}, fmt.Errorf("no ABI file or directory provided")
	}

	specs := make([]*abi.AbiSpec, 0)

	err := filepath.Walk(abiFileOrDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error returned while walking abiDir '%s': %v", abiFileOrDir, err)
		}
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
	return abi.MergeAbiSpec(specs), nil
}
