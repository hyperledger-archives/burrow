package tm

import (
	"context"
	"fmt"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/txs"
	gorpc "github.com/tendermint/tendermint/rpc/lib/server"
	"github.com/tendermint/tendermint/rpc/lib/types"
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

const SubscriptionTimeoutSeconds = 5 * time.Second

func GetRoutes(service *rpc.Service, logger logging_types.InfoTraceLogger) map[string]*gorpc.RPCFunc {
	logger = logging.WithScope(logger, "GetRoutes")
	return map[string]*gorpc.RPCFunc{
		// Transact
		BroadcastTx: gorpc.NewRPCFunc(func(tx txs.Wrapper) (*rpc.ResultBroadcastTx, error) {
			receipt, err := service.Transactor().BroadcastTx(tx.Unwrap())
			if err != nil {
				return nil, err
			}
			return &rpc.ResultBroadcastTx{
				Receipt: *receipt,
			}, nil
		}, "tx"),

		SignTx: gorpc.NewRPCFunc(func(tx txs.Tx, concretePrivateAccounts []*acm.ConcretePrivateAccount) (*rpc.ResultSignTx, error) {
			tx, err := service.Transactor().SignTx(tx, acm.PrivateAccounts(concretePrivateAccounts))
			return &rpc.ResultSignTx{Tx: txs.Wrap(tx)}, err

		}, "tx,privAccounts"),

		// Simulated call
		Call: gorpc.NewRPCFunc(func(fromAddress, toAddress acm.Address, data []byte) (*rpc.ResultCall, error) {
			call, err := service.Transactor().Call(fromAddress, toAddress, data)
			if err != nil {
				return nil, err
			}
			return &rpc.ResultCall{Call: *call}, nil
		}, "fromAddress,toAddress,data"),

		CallCode: gorpc.NewRPCFunc(func(fromAddress acm.Address, code, data []byte) (*rpc.ResultCall, error) {
			call, err := service.Transactor().CallCode(fromAddress, code, data)
			if err != nil {
				return nil, err
			}
			return &rpc.ResultCall{Call: *call}, nil
		}, "fromAddress,code,data"),

		// Events
		Subscribe: gorpc.NewWSRPCFunc(func(wsCtx rpctypes.WSRPCContext, eventID string) (*rpc.ResultSubscribe, error) {
			subscriptionID, err := event.GenerateSubscriptionID()
			if err != nil {
				return nil, err
			}
			ctx, cancel := context.WithTimeout(context.Background(), SubscriptionTimeoutSeconds*time.Second)
			defer cancel()
			err = service.Subscribe(ctx, subscriptionID, eventID, func(resultEvent *rpc.ResultEvent) bool {
				keepAlive := wsCtx.TryWriteRPCResponse(rpctypes.NewRPCSuccessResponse(
					EventResponseID(wsCtx.Request.ID, eventID), resultEvent))
				if !keepAlive {
					logging.InfoMsg(logger, "dropping subscription because could not write to websocket",
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

		Unsubscribe: gorpc.NewWSRPCFunc(func(wsCtx rpctypes.WSRPCContext, subscriptionID string) (*rpc.ResultUnsubscribe, error) {
			ctx, cancel := context.WithTimeout(context.Background(), SubscriptionTimeoutSeconds*time.Second)
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
		Status:  gorpc.NewRPCFunc(service.Status, ""),
		NetInfo: gorpc.NewRPCFunc(service.NetInfo, ""),

		// Accounts
		ListAccounts: gorpc.NewRPCFunc(func() (*rpc.ResultListAccounts, error) {
			return service.ListAccounts(func(acm.Account) bool {
				return true
			})
		}, ""),

		GetAccount:      gorpc.NewRPCFunc(service.GetAccount, "address"),
		GetStorage:      gorpc.NewRPCFunc(service.GetStorage, "address,key"),
		DumpStorage:     gorpc.NewRPCFunc(service.DumpStorage, "address"),
		GetAccountHuman: gorpc.NewRPCFunc(service.GetAccountHumanReadable, "address"),

		// Blockchain
		Genesis:    gorpc.NewRPCFunc(service.Genesis, ""),
		ChainID:    gorpc.NewRPCFunc(service.ChainId, ""),
		ListBlocks: gorpc.NewRPCFunc(service.ListBlocks, "minHeight,maxHeight"),
		GetBlock:   gorpc.NewRPCFunc(service.GetBlock, "height"),

		// Consensus
		ListUnconfirmedTxs: gorpc.NewRPCFunc(service.ListUnconfirmedTxs, "maxTxs"),
		ListValidators:     gorpc.NewRPCFunc(service.ListValidators, ""),
		DumpConsensusState: gorpc.NewRPCFunc(service.DumpConsensusState, ""),

		// Names
		GetName: gorpc.NewRPCFunc(service.GetName, "name"),
		ListNames: gorpc.NewRPCFunc(func() (*rpc.ResultListNames, error) {
			return service.ListNames(func(*execution.NameRegEntry) bool {
				return true
			})
		}, ""),

		// Private account
		GeneratePrivateAccount: gorpc.NewRPCFunc(service.GeneratePrivateAccount, ""),
	}
}

// In a slight abuse of the JSON-RPC spec (it states we should return the same ID as provided by the client)
// we append the eventID to the websocket response ID when pushing events over
// the websocket to distinguish events themselves from the initial ResultSubscribe response.
func EventResponseID(requestID, eventID string) string {
	return fmt.Sprintf("%s#%s", requestID, eventID)
}
