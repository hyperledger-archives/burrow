package jobs

import (
	"github.com/hyperledger/burrow/deploy/def"
	log "github.com/sirupsen/logrus"
)

func SetAccountJob(account *def.Account, do *def.DeployArgs, script *def.Playbook) (string, error) {
	var result string

	// Set the Account in the Package & Announce
	script.Account = account.Address
	log.WithField("=>", script.Account).Info("Setting Account")

	// Set result and return
	result = account.Address
	return result, nil
}

func SetValJob(set *def.Set, do *def.DeployArgs) (string, error) {
	var result string
	log.WithField("=>", set.Value).Info("Setting Variable")
	result = set.Value
	return result, nil
}
