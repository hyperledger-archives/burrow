package loader

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/logging"
	"github.com/spf13/viper"
)

func LoadPlaybook(fileName string, args *def.DeployArgs, logger *logging.Logger) (*def.Playbook, error) {
	return loadPlaybook(fileName, args, nil, logger)
}

func loadPlaybook(fileName string, args *def.DeployArgs, parent *def.Playbook, logger *logging.Logger) (*def.Playbook, error) {
	logger.InfoMsg("Loading Playbook File.")
	playbook := new(def.Playbook)
	deployJobs := viper.New()

	if parent != nil {
		// if subYAMLPath does not exist, try YAMLPath relative to paremt.Path
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			fileName = filepath.Join(parent.Path, fileName)
			logger.InfoMsg("Trying Playbook relative to Path", "filename", fileName, "path", parent.Path)
		}
	}

	playbook.Filename = fileName
	playbook.Path = filepath.Dir(fileName)

	if args.BinPath == "[dir]/bin" {
		playbook.BinPath = filepath.Join(playbook.Path, "bin")
	} else {
		playbook.BinPath = args.BinPath
	}

	if err := os.MkdirAll(playbook.BinPath, 0775); err != nil {
		return nil, err
	}

	// setup file
	abs, err := filepath.Abs(fileName)
	if err != nil {
		return nil, fmt.Errorf("sorry, the marmots were unable to find the absolute path to the playbook file")
	}

	base := filepath.Base(abs)
	extName := filepath.Ext(base)
	bName := base[:len(base)-len(extName)]
	logger.InfoMsg("Loading playbook file",
		"path", playbook.Path,
		"filename", playbook.Filename,
	)

	deployJobs.SetConfigType("yaml")
	deployJobs.SetConfigName(bName)

	r, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	// load file
	if err := deployJobs.ReadConfig(r); err != nil {
		return nil, fmt.Errorf("Sorry, the marmots were unable to load the playbook file. Please check your path: %v", err)
	}

	// marshall file
	if err := deployJobs.UnmarshalExact(playbook); err != nil {
		return nil, fmt.Errorf(`Sorry, the marmots could not figure that playbook file out.
			Please check that your deploy.yaml is properly formatted: %v`, err)
	}

	// TODO more file sanity check (fail before running)
	err = playbook.Validate()
	if err != nil {
		return nil, err
	}

	for _, job := range playbook.Jobs {
		if job.Meta != nil {
			metaPlaybook, err := loadPlaybook(job.Meta.File, args, playbook, logger)
			if err != nil {
				return nil, err
			}

			// set the deploy contract jobs relative to the newDo's root directory
			for _, job := range metaPlaybook.Jobs {
				if job.Deploy != nil {
					job.Deploy.Contract = filepath.Join(metaPlaybook.Path, job.Deploy.Contract)
				}
			}

			// We do not set the parent for this playbook; the parent is used for
			// backreferencing variables
			job.Meta.Playbook = metaPlaybook
		}

		if job.Proposal != nil {
			for _, job := range job.Proposal.Jobs {
				if job.Meta != nil {
					metaPlaybook, err := loadPlaybook(job.Meta.File, args, playbook, logger)
					if err != nil {
						return nil, err
					}

					// set the deploy contract jobs relative to the newDo's root directory
					for _, job := range metaPlaybook.Jobs {
						if job.Deploy != nil {
							job.Deploy.Contract = filepath.Join(metaPlaybook.Path, job.Deploy.Contract)
						}
					}

					// Set the parent for the playbook so that the proposal can backreference e.g.
					// deployed contracts addresses
					metaPlaybook.Parent = playbook
					job.Meta.Playbook = metaPlaybook
				}
			}
		}
	}

	return playbook, nil
}
