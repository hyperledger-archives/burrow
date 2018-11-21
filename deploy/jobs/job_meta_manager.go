package jobs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/loader"
	log "github.com/sirupsen/logrus"
)

func MetaJob(meta *def.Meta, do *def.DeployArgs, client *def.Client) (string, error) {
	var err error
	var pwd string

	pwd, err = os.Getwd()
	if err != nil {
		return "failed", err
	}

	// work from a fresh Do object
	newDo := new(def.DeployArgs)
	newDo.Address = do.Address
	newDo.CurrentOutput = do.CurrentOutput
	newDo.DefaultAmount = do.DefaultAmount
	newDo.DefaultFee = do.DefaultFee
	newDo.DefaultGas = do.DefaultGas
	newDo.DefaultSets = do.DefaultSets
	newDo.Jobs = do.Jobs

	// Set subYAMLPath
	newDo.YAMLPath = meta.File
	log.WithField("=>", newDo.YAMLPath).Info("Trying base YAMLPath")

	// if subYAMLPath does not exist, try YAMLPath relative to do.Path
	if _, err := os.Stat(newDo.YAMLPath); os.IsNotExist(err) {
		newDo.YAMLPath = filepath.Join(do.Path, newDo.YAMLPath)
		log.WithField("=>", newDo.YAMLPath).Info("Trying YAMLPath relative to do.Path")
	}

	// if subYAMLPath does not exist, try YAMLPath relative to pwd
	if _, err := os.Stat(newDo.YAMLPath); os.IsNotExist(err) {
		newDo.YAMLPath = filepath.Join(pwd, newDo.YAMLPath)
		log.WithField("=>", newDo.YAMLPath).Info("Trying YAMLPath relative to pwd")
	}

	// if subYAMLPath cannot be found, abort
	if _, err := os.Stat(newDo.YAMLPath); os.IsNotExist(err) {
		return "failed", fmt.Errorf("could not find sub YAML file (%s)",
			newDo.YAMLPath)
	}

	// once we have the proper subYAMLPath set the paths accordingly
	newDo.Path = filepath.Dir(newDo.YAMLPath)
	newDo.BinPath = filepath.Join(newDo.Path, filepath.Base(do.BinPath))

	// load the package
	log.WithField("=>", newDo.YAMLPath).Info("Loading sub YAML")
	script, err := loader.LoadPackage(newDo.YAMLPath)
	if err != nil {
		return "failed", err
	}

	// set the deploy contract jobs relative to the newDo's root directory
	for _, job := range script.Jobs {
		if job.Deploy != nil {
			job.Deploy.Contract = filepath.Join(newDo.Path, job.Deploy.Contract)
		}
	}

	err = DoJobs(newDo, script, client)
	if err != nil {
		return "failed", err
	}

	do.CurrentOutput = ""
	return "passed", nil
}
