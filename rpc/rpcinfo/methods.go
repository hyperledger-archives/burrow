package rpcinfo

import (
	"fmt"
	"regexp"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/lib/server"
)

// Method names
const (
	// Status and healthcheck
	Status  = "status"
	Network = "network"

	// Accounts
	Accounts        = "accounts"
	Account         = "account"
	Storage         = "storage"
	DumpStorage     = "dump_storage"
	GetAccountHuman = "account_human"
	AccountStats    = "account_stats"

	// Names
	Name  = "name"
	Names = "names"

	// Blockchain
	Genesis = "genesis"
	ChainID = "chain_id"
	Block   = "block"
	Blocks  = "blocks"

	// Consensus
	UnconfirmedTxs = "unconfirmed_txs"
	Validators     = "validators"
	Consensus      = "consensus"
)

const maxRegexLength = 255

// The methods below all get mounted at the info server address (specified in config at RPC/Info) in the following form:
//
// http://<info-host>:<info-port>/<name>?<param1>=<value1>&<param2>=<value2>[&...]
//
// For example:
// http://0.0.0.0:26658/status?block_time_within=10m&block_seen_time_within=1h
// http://0.0.0.0:26658/names?regex=<regular expression to match name>
//
// They keys in the route map below are the endpoint name, and the comma separated values are the url query params
//
// They info endpoint also all be called with a JSON-RPC payload like:
//
// curl -X POST -d '{"method": "names", "id": "foo", "params": ["loves"]}' http://0.0.0.0:26658
//
func GetRoutes(service *rpc.Service) map[string]*server.RPCFunc {
	// TODO: overhaul this with gRPC-gateway / swagger
	return map[string]*server.RPCFunc{
		// Status
		Status:  server.NewRPCFunc(service.StatusWithin, "block_time_within,block_seen_time_within"),
		Network: server.NewRPCFunc(service.Network, ""),

		// Accounts
		Accounts: server.NewRPCFunc(func() (*rpc.ResultAccounts, error) {
			return service.Accounts(func(*acm.Account) bool {
				return true
			})
		}, ""),

		Account:         server.NewRPCFunc(service.Account, "address"),
		Storage:         server.NewRPCFunc(service.Storage, "address,key"),
		DumpStorage:     server.NewRPCFunc(service.DumpStorage, "address"),
		GetAccountHuman: server.NewRPCFunc(service.AccountHumanReadable, "address"),
		AccountStats:    server.NewRPCFunc(service.AccountStats, ""),

		// Blockchain
		Genesis: server.NewRPCFunc(service.Genesis, ""),
		ChainID: server.NewRPCFunc(service.ChainIdentifiers, ""),
		Blocks:  server.NewRPCFunc(service.Blocks, "minHeight,maxHeight"),
		Block:   server.NewRPCFunc(service.Block, "height"),

		// Consensus
		UnconfirmedTxs: server.NewRPCFunc(service.UnconfirmedTxs, "maxTxs"),
		Validators:     server.NewRPCFunc(service.Validators, ""),
		Consensus:      server.NewRPCFunc(service.ConsensusState, ""),

		// Names
		Name: server.NewRPCFunc(service.Name, "name"),
		Names: server.NewRPCFunc(func(regex string) (*rpc.ResultNames, error) {
			if regex == "" {
				return service.Names(func(*names.Entry) bool { return true })
			}
			// Regex attacks...
			if len(regex) > maxRegexLength {
				return nil, fmt.Errorf("regular expression longer than maximum length %d", maxRegexLength)
			}
			re, err := regexp.Compile(regex)
			if err != nil {
				return nil, fmt.Errorf("could not compile '%s' as regular expression: %v", regex, err)
			}
			return service.Names(func(entry *names.Entry) bool {
				return re.MatchString(entry.Name)
			})
		}, "regex"),
	}
}
