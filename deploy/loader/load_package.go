package loader

import (
	"fmt"
	"path/filepath"

	"os"

	"github.com/hyperledger/burrow/deploy/def"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func LoadPackage(fileName string) (*def.Package, error) {
	log.Info("Loading monax Jobs Definition File.")
	var pkg = new(def.Package)
	var deployJobs = viper.New()

	// setup file
	abs, err := filepath.Abs(fileName)
	if err != nil {
		return nil, fmt.Errorf("Sorry, the marmots were unable to find the absolute path to the monax jobs file.")
	}

	dir := filepath.Dir(abs)
	base := filepath.Base(abs)
	extName := filepath.Ext(base)
	bName := base[:len(base)-len(extName)]
	log.WithFields(log.Fields{
		"path": dir,
		"name": bName,
	}).Debug("Loading jobs file")

	deployJobs.SetConfigType("yaml")
	deployJobs.SetConfigName(bName)

	r, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	// load file
	if err := deployJobs.ReadConfig(r); err != nil {
		return nil, fmt.Errorf("Sorry, the marmots were unable to load the monax jobs file. Please check your path: %v", err)
	}

	// marshall file
	if err := deployJobs.UnmarshalExact(pkg); err != nil {
		return nil, fmt.Errorf(`Sorry, the marmots could not figure that monax jobs file out.
			Please check that your deploy.yaml is properly formatted: %v`, err)
	}

	// TODO more file sanity check (fail before running)

	return pkg, nil
}
