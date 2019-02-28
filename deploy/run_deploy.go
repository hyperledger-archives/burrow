package pkgs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/jobs"
	"github.com/hyperledger/burrow/deploy/loader"
	"github.com/hyperledger/burrow/execution/evm/abi"
	log "github.com/sirupsen/logrus"
)

func RunPlaybook(args *def.DeployArgs, playbooks []string) error {
	var err error
	var pwd string

	pwd, err = os.Getwd()
	if err != nil {
		return err
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

	for _, playbook := range playbooks {
		// block that triggers if the do.Path was NOT set
		//   via cli flag... or not
		if args.Path == "" {
			// if do.YAMLPath does not exist, try YAMLPath relative to pwd
			if _, err := os.Stat(playbook); os.IsNotExist(err) {
				playbook = filepath.Join(pwd, playbook)
			}
		} else {
			// --dir is given, assume YAMLPath relative to dirPath
			playbook = filepath.Join(args.Path, playbook)

			// if do.YAMLPath does not exist, try YAMLPath relative to pwd
			if _, err := os.Stat(playbook); os.IsNotExist(err) {
				playbook = filepath.Join(pwd, playbook)
			}
		}

		// if YAMLPath cannot be found, abort
		if _, err := os.Stat(playbook); os.IsNotExist(err) {
			return fmt.Errorf("could not find playbook file (%s)",
				playbook)
		}

		client := def.NewClient(args.Chain, args.KeysService, args.MempoolSign, time.Duration(args.Timeout)*time.Second)
		client.AllSpecs, err = abi.LoadPath(args.BinPath)
		if err != nil {
			log.Errorf("failed to load ABIs for Event parsing from %s: %v", args.BinPath, err)
		}

		// useful for debugging
		printPathPackage(client)

		// Load the package if it doesn't exist
		script, err := loader.LoadPlaybook(playbook, args)
		if err != nil {
			return err
		}

		// Ensure relative paths if we're given a different path for deploy contracts jobs
		// Solidity contracts may import other solidity contracts, and the working directory
		// is the directory where solc searches from.
		if args.Path != "" && args.Path != pwd {
			err = os.Chdir(args.Path)
			if err != nil {
				return err
			}
		}

		err = jobs.ExecutePlaybook(args, script, client)
		if err != nil {
			return err
		}
	}

	return nil
}

func printPathPackage(client *def.Client) {
	log.WithField("=>", client.ChainAddress).Info("With ChainURL")
	log.WithField("=>", client.KeysClientAddress).Info("Using Signer at")
}
