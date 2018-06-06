package tm

import (
	"context"
	"fmt"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/tm/lib/server"
	"github.com/hyperledger/burrow/rpc/tm/lib/types"
	"github.com/hyperledger/burrow/txs"
)

// Method names
const (
	BroadcastTx = "broadcast_tx"
	Subscribe   = "subscribe"
	Unsubscribe = "unsubscribe"

	// Status
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

const SubscriptionTimeout = 5 * time.Second

func GetRoutes(service *rpc.Service, logger *logging.Logger) map[string]*server.RPCFunc {
	logger = logger.WithScope("GetRoutes")
	return map[string]*server.RPCFunc{
		// Transact
		BroadcastTx: server.NewRPCFunc(func(tx txs.Body) (*rpc.ResultBroadcastTx, error) {
			receipt, err := service.Transactor().BroadcastTx(tx.Unwrap())
			if err != nil {
				return nil, err
			}
			return &rpc.ResultBroadcastTx{
				Receipt: *receipt,
			}, nil
		}, "tx"),

		SignTx: server.NewRPCFunc(func(tx txs.Tx, concretePrivateAccounts []*acm.ConcretePrivateAccount) (*rpc.ResultSignTx, error) {
			tx, err := service.Transactor().SignTx(tx, acm.SigningAccounts(concretePrivateAccounts))
			return &rpc.ResultSignTx{Tx: txs.Wrap(tx)}, err

		}, "tx,privAccounts"),

		// Simulated call
		Call: server.NewRPCFunc(func(fromAddress, toAddress crypto.Address, data []byte) (*rpc.ResultCall, error) {
			call, err := service.Transactor().Call(service.State(), fromAddress, toAddress, data)
			if err != nil {
				return nil, err
			}
			return &rpc.ResultCall{Call: *call}, nil
		}, "fromAddress,toAddress,data"),

		CallCode: server.NewRPCFunc(func(fromAddress crypto.Address, code, data []byte) (*rpc.ResultCall, error) {
			call, err := service.Transactor().CallCode(service.State(), fromAddress, code, data)
			if err != nil {
				return nil, err
			}
			return &rpc.ResultCall{Call: *call}, nil
		}, "fromAddress,code,data"),

		// Events
		Subscribe: server.NewWSRPCFunc(func(wsCtx types.WSRPCContext, eventID string) (*rpc.ResultSubscribe, error) {
			subscriptionID, err := event.GenerateSubscriptionID()
			if err != nil {
				return nil, err
			}

			ctx, cancel := context.WithTimeout(context.Background(), SubscriptionTimeout)
			defer cancel()

			err = service.Subscribe(ctx, subscriptionID, eventID, func(resultEvent *rpc.ResultEvent) bool {
				keepAlive := wsCtx.TryWriteRPCResponse(types.NewRPCSuccessResponse(
					EventResponseID(wsCtx.Request.ID, eventID), resultEvent))
				if !keepAlive {
					logger.InfoMsg("dropping subscription because could not write to websocket",
						"subscription_id", subscriptionID,
						"event_id", eventID)
				}
				return keepAlive
			})
			if err != nil {
				return nil, err
			}
			return &rpc.ResultSubscribe{
				EventID:        eventID,
				SubscriptionID: subscriptionID,
			}, nil
		}, "eventID"),

		Unsubscribe: server.NewWSRPCFunc(func(wsCtx types.WSRPCContext, subscriptionID string) (*rpc.ResultUnsubscribe, error) {
			ctx, cancel := context.WithTimeout(context.Background(), SubscriptionTimeout)
			defer cancel()
			// Since our model uses a random subscription ID per request we just drop all matching requests
			err := service.Unsubscribe(ctx, subscriptionID)
			if err != nil {
				return nil, err
			}
			return &rpc.ResultUnsubscribe{
				SubscriptionID: subscriptionID,
			}, nil
		}, "subscriptionID"),

		// Status
		Status:  server.NewRPCFunc(service.Status, ""),
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
		ChainID:    server.NewRPCFunc(service.ChainId, ""),
		ListBlocks: server.NewRPCFunc(service.ListBlocks, "minHeight,maxHeight"),
		GetBlock:   server.NewRPCFunc(service.GetBlock, "height"),

		// Consensus
		ListUnconfirmedTxs: server.NewRPCFunc(service.ListUnconfirmedTxs, "maxTxs"),
		ListValidators:     server.NewRPCFunc(service.ListValidators, ""),
		DumpConsensusState: server.NewRPCFunc(service.DumpConsensusState, ""),

		// Names
		GetName: server.NewRPCFunc(service.GetName, "name"),
		ListNames: server.NewRPCFunc(func() (*rpc.ResultListNames, error) {
			return service.ListNames(func(*execution.NameRegEntry) bool {
				return true
			})
		}, ""),

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
