package util

import (
	"fmt"
	"os"

	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/logging"
)

func Exit(err error) {
	status := 0
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		status = 1
	}
	os.Exit(status)
}

func ChainErrorHandler(account string, err error, logger *logging.Logger) error {
	logger.InfoMsg("There has been an error talking to your Burrow chain",
		"defAddr", account,
		"rawErr", err)

	return fmt.Errorf(`
There has been an error talking to your Burrow chain using account %s.

%v

`, account, err)
}

func ABIErrorHandler(err error, call *def.Call, query *def.QueryContract, logger *logging.Logger) error {
	switch {
	case call != nil:
		logger.InfoMsg("ABI Error",
			"data", call.Data,
			"bin", call.Bin,
			"dest", call.Destination,
			"rawErr", err)
	case query != nil:
		logger.InfoMsg("ABI Error",
			"data", query.Data,
			"bin", query.Bin,
			"dest", query.Destination,
			"rawErr", err)
	}

	return fmt.Errorf(`
There has been an error in finding or in using your ABI. ABI's are "Application Binary
Interface" and they are what let us know how to talk to smart contracts.

These little json files can be read by a variety of things which need to talk to smart
contracts so they are quite necessary to be able to find and use properly.

The ABIs are saved after the deploy events. So if there was a glitch in the matrix,
we apologize in advance.

The marmot recovery checklist is...
  * ensure your chain is running and you have enough validators online
  * ensure that your contracts successfully deployed
  * if you used imports or have multiple contracts in one file check the instance
    variable in the deploy and the abi variable in the call/query-contract
  * make sure you're calling or querying the right function
  * make sure you're using the correct variables for job results
`)
}
