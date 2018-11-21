package pkgs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/jobs"
	"github.com/hyperledger/burrow/deploy/loader"
	log "github.com/sirupsen/logrus"
)

func RunPackage(do *def.DeployArgs, script *def.Playbook, client *def.Client) error {
	var err error
	var pwd string

	pwd, err = os.Getwd()
	if err != nil {
		return err
	}
	originalYAMLPath := do.YAMLPath

	// block that triggers if the do.Path was NOT set
	//   via cli flag... or not
	if do.Path == "" {
		do.Path = pwd

		// if do.YAMLPath does not exist, try YAMLPath relative to pwd
		if _, err := os.Stat(do.YAMLPath); os.IsNotExist(err) {
			do.YAMLPath = filepath.Join(pwd, originalYAMLPath)
		}
	} else {
		// --dir is given, assume YAMLPath relative to dirPath
		do.YAMLPath = filepath.Join(do.Path, originalYAMLPath)

		// if do.YAMLPath does not exist, try YAMLPath relative to pwd
		if _, err := os.Stat(do.YAMLPath); os.IsNotExist(err) {
			do.YAMLPath = filepath.Join(pwd, originalYAMLPath)
		}
	}

	// if YAMLPath cannot be found, abort
	if _, err := os.Stat(do.YAMLPath); os.IsNotExist(err) {
		return fmt.Errorf("could not find jobs file (%s), ensure correct used of the --file flag",
			do.YAMLPath)
	}

	// if bin and abi paths are default cli settings then use the
	//   stated defaults of do.Path plus bin|abi
	if do.BinPath == "[dir]/bin" {
		do.BinPath = filepath.Join(do.Path, "bin")
	}

	// useful for debugging
	printPathPackage(client)

	// Load the package if it doesn't exist
	if script == nil {
		script, err = loader.LoadPackage(do.YAMLPath)
		if err != nil {
			return err
		}
	}

	// Ensure relative paths if we're given a different path for deploy contracts jobs
	// Solidity contracts may import other solidity contracts, and the working directory
	// is the directory where solc searches from.
	if do.Path != pwd {
		err = os.Chdir(do.Path)
		if err != nil {
			return err
		}
	}

	return jobs.DoJobs(do, script, client)
}

func printPathPackage(client *def.Client) {
	log.WithField("=>", client.ChainAddress).Info("With ChainURL")
	log.WithField("=>", client.KeysClientAddress).Info("Using Signer at")
}
