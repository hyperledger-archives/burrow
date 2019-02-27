package pkgs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/jobs"
	"github.com/hyperledger/burrow/deploy/loader"
	"github.com/hyperledger/burrow/execution/evm/abi"
	log "github.com/sirupsen/logrus"
)

func RunPackage(args *def.DeployArgs, playbooks []string, client *def.Client) error {
	var err error
	var pwd string

	pwd, err = os.Getwd()
	if err != nil {
		return err
	}
	originalYAMLPath := args.YAMLPath

	// block that triggers if the do.Path was NOT set
	//   via cli flag... or not
	if args.Path == "" {
		args.Path = pwd

		// if do.YAMLPath does not exist, try YAMLPath relative to pwd
		if _, err := os.Stat(args.YAMLPath); os.IsNotExist(err) {
			args.YAMLPath = filepath.Join(pwd, originalYAMLPath)
		}
	} else {
		// --dir is given, assume YAMLPath relative to dirPath
		args.YAMLPath = filepath.Join(args.Path, originalYAMLPath)

		// if do.YAMLPath does not exist, try YAMLPath relative to pwd
		if _, err := os.Stat(args.YAMLPath); os.IsNotExist(err) {
			args.YAMLPath = filepath.Join(pwd, originalYAMLPath)
		}
	}

	// if YAMLPath cannot be found, abort
	if _, err := os.Stat(args.YAMLPath); os.IsNotExist(err) {
		return fmt.Errorf("could not find jobs file (%s), ensure correct used of the --file flag",
			args.YAMLPath)
	}

	// if bin and abi paths are default cli settings then use the
	//   stated defaults of do.Path plus bin|abi
	if args.BinPath == "[dir]/bin" {
		args.BinPath = filepath.Join(args.Path, "bin")
	}

	if _, err := os.Stat(args.BinPath); os.IsNotExist(err) {
		if err := os.Mkdir(args.BinPath, 0775); err != nil {
			return err
		}
	}

	client.AllSpecs, err = abi.LoadPath(args.BinPath)
	if err != nil {
		log.Errorf("failed to load ABIs for Event parsing from %s: %v", args.BinPath, err)
	}

	// useful for debugging
	printPathPackage(client)

	// Load the package if it doesn't exist
	script, err := loader.LoadPlaybook(args.YAMLPath, args)
	if err != nil {
		return err
	}

	// Ensure relative paths if we're given a different path for deploy contracts jobs
	// Solidity contracts may import other solidity contracts, and the working directory
	// is the directory where solc searches from.
	if args.Path != pwd {
		err = os.Chdir(args.Path)
		if err != nil {
			return err
		}
	}

	return jobs.ExecutePlaybook(args, script, client)
}

func printPathPackage(client *def.Client) {
	log.WithField("=>", client.ChainAddress).Info("With ChainURL")
	log.WithField("=>", client.KeysClientAddress).Info("Using Signer at")
}
