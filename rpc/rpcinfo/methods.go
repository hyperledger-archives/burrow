package rpcinfo

import (
	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/logging"
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

	// Simulated call
	Call     = "call"
	CallCode = "call_code"

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

func GetRoutes(service *rpc.Service, logger *logging.Logger) map[string]*server.RPCFunc {
	logger = logger.WithScope("GetRoutes")
	return map[string]*server.RPCFunc{
		// Status
		Status:  server.NewRPCFunc(service.StatusWithin, "block_time_within,block_seen_time_within"),
		Network: server.NewRPCFunc(service.Network, ""),

		// Accounts
		Accounts: server.NewRPCFunc(func() (*rpc.ResultAccounts, error) {
			return service.Accounts(func(acm.Account) bool {
				return true
			})
		}, ""),

		Account:         server.NewRPCFunc(service.Account, "address"),
		Storage:         server.NewRPCFunc(service.Storage, "address,key"),
		DumpStorage:     server.NewRPCFunc(service.DumpStorage, "address"),
		GetAccountHuman: server.NewRPCFunc(service.AccountHumanReadable, "address"),

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
		Name:  server.NewRPCFunc(service.Name, "name"),
		Names: server.NewRPCFunc(service.Names, ""),
	}
}
