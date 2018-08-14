package jobs

import (
	"github.com/hyperledger/burrow/deploy/def"
	log "github.com/sirupsen/logrus"
)

func SetAccountJob(account *def.Account, do *def.Packages) (string, error) {
	var result string

	// Set the Account in the Package & Announce
	do.Package.Account = account.Address
	log.WithField("=>", do.Package.Account).Info("Setting Account")

	// Set result and return
	result = account.Address
	return result, nil
}

func SetValJob(set *def.Set, do *def.Packages) (string, error) {
	var result string
	log.WithField("=>", set.Value).Info("Setting Variable")
	result = set.Value
	return result, nil
}
