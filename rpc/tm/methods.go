package tm

import (
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/lib/server"
)

// Method names
const (
	// Status and healthcheck
	Status  = "status"
	NetInfo = "net_info"

	// Accounts
	ListAccounts    = "list_accounts"
	GetAccount      = "get_account"
	GetStorage      = "get_storage"
	DumpStorage     = "dump_storage"
	GetAccountHuman = "get_account_human"

	// Simulated call
	Call     = "call"
	CallCode = "call_code"

	// Names
	GetName   = "get_name"
	ListNames = "list_names"

	// Blockchain
	Genesis    = "genesis"
	ChainID    = "chain_id"
	GetBlock   = "get_block"
	ListBlocks = "list_blocks"

	// Consensus
	ListUnconfirmedTxs = "list_unconfirmed_txs"
	ListValidators     = "list_validators"
	DumpConsensusState = "dump_consensus_state"

	// Private keys and signing
	GeneratePrivateAccount = "unsafe/gen_priv_account"
	SignTx                 = "unsafe/sign_tx"
)

func GetRoutes(service *rpc.Service, logger *logging.Logger) map[string]*server.RPCFunc {
	logger = logger.WithScope("GetRoutes")
	return map[string]*server.RPCFunc{
		// Status
		Status:  server.NewRPCFunc(service.Status, "block_within"),
		NetInfo: server.NewRPCFunc(service.NetInfo, ""),

		// Accounts
		ListAccounts: server.NewRPCFunc(func() (*rpc.ResultListAccounts, error) {
			return service.ListAccounts(func(acm.Account) bool {
				return true
			})
		}, ""),

		GetAccount:      server.NewRPCFunc(service.GetAccount, "address"),
		GetStorage:      server.NewRPCFunc(service.GetStorage, "address,key"),
		DumpStorage:     server.NewRPCFunc(service.DumpStorage, "address"),
		GetAccountHuman: server.NewRPCFunc(service.GetAccountHumanReadable, "address"),

		// Blockchain
		Genesis:    server.NewRPCFunc(service.Genesis, ""),
		ChainID:    server.NewRPCFunc(service.ChainIdentifiers, ""),
		ListBlocks: server.NewRPCFunc(service.ListBlocks, "minHeight,maxHeight"),
		GetBlock:   server.NewRPCFunc(service.GetBlock, "height"),

		// Consensus
		ListUnconfirmedTxs: server.NewRPCFunc(service.ListUnconfirmedTxs, "maxTxs"),
		ListValidators:     server.NewRPCFunc(service.ListValidators, ""),
		DumpConsensusState: server.NewRPCFunc(service.DumpConsensusState, ""),

		// Names
		GetName:   server.NewRPCFunc(service.GetName, "name"),
		ListNames: server.NewRPCFunc(service.ListNames, ""),

		// Private account
		GeneratePrivateAccount: server.NewRPCFunc(service.GeneratePrivateAccount, ""),
	}
}

// In a slight abuse of the JSON-RPC spec (it states we should return the same ID as provided by the client)
// we append the eventID to the websocket response ID when pushing events over
// the websocket to distinguish events themselves from the initial ResultSubscribe response.
func EventResponseID(requestID, eventID string) string {
	return fmt.Sprintf("%s#%s", requestID, eventID)
}
