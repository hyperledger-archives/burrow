package ethclient

import (
	"context"
	"fmt"
	"time"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/web3"
)

const (
	EthGetLogsMethod               = "eth_getLogs"
	EthSyncingMethod               = "eth_syncing"
	EthBlockNumberMethod           = "eth_blockNumber"
	EthSendTransactionMethod       = "eth_sendTransaction"
	EthSendRawTransactionMethod    = "eth_sendRawTransaction"
	EthGetTransactionCountMethod   = "eth_getTransactionCount"
	EthAccountsMethod              = "eth_accounts"
	EthGetTransactionByHashMethod  = "eth_getTransactionByHash"
	EthGetTransactionReceiptMethod = "eth_getTransactionReceipt"
	EthGasPriceMethod              = "eth_gasPrice"
	NetVersionMethod               = "net_version"
	Web3ClientVersionMethod        = "web3_clientVersion"
)

// Adjust the polling frequency of AwaitTransaction
const awaitTransactionSleep = 200 * time.Millisecond

func EthSendTransaction(client rpc.Client, tx *EthSendTransactionParam) (string, error) {
	hash := new(string)
	err := client.Call(EthSendTransactionMethod, []*EthSendTransactionParam{tx}, &hash)
	if err != nil {
		return "", err
	}
	return *hash, nil
}

func EthSendRawTransaction(client rpc.Client, txHex string) (string, error) {
	hash := new(string)
	err := client.Call(EthSendRawTransactionMethod, []string{txHex}, &hash)
	if err != nil {
		return "", err
	}
	return *hash, nil
}

func EthGetTransactionCount(client rpc.Client, address crypto.Address) (string, error) {
	var count string
	err := client.Call(EthGetTransactionCountMethod, []string{web3.HexEncoder.Address(address), "latest"}, &count)
	if err != nil {
		return "", err
	}
	return count, nil
}

func EthGetLogs(client rpc.Client, filter *Filter) ([]*EthLog, error) {
	var logs []*EthLog
	err := client.Call(EthGetLogsMethod, []*EthFilter{filter.EthFilter()}, &logs)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func EthAccounts(client rpc.Client) ([]string, error) {
	var accounts []string
	err := client.Call(EthAccountsMethod, nil, &accounts)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func EthGetTransactionByHash(client rpc.Client, txHash string) (*web3.Transaction, error) {
	tx := new(web3.Transaction)
	err := client.Call(EthGetTransactionByHashMethod, []string{txHash}, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func EthGetTransactionReceipt(client rpc.Client, txHash string) (*Receipt, error) {
	tx := new(Receipt)
	err := client.Call(EthGetTransactionReceiptMethod, []string{txHash}, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func EthSyncing(client rpc.Client) (bool, error) {
	syncing := new(bool)
	err := client.Call(EthSyncingMethod, nil, syncing)
	if err != nil {
		return false, err
	}
	return *syncing, nil
}

func EthBlockNumber(client rpc.Client) (uint64, error) {
	latestBlock := new(string)
	err := client.Call(EthBlockNumberMethod, nil, latestBlock)
	if err != nil {
		return 0, err
	}
	d := new(web3.HexDecoder)
	return d.Uint64(*latestBlock), d.Err()
}

func EthGasPrice(client rpc.Client) (string, error) {
	gasPrice := new(string)
	err := client.Call(EthGasPriceMethod, nil, gasPrice)
	if err != nil {
		return "", err
	}
	return *gasPrice, nil
}

// AKA ChainID
func NetVersion(client rpc.Client) (string, error) {
	version := new(string)
	err := client.Call(NetVersionMethod, nil, version)
	if err != nil {
		return "", err
	}
	return *version, nil
}

func Web3ClientVersion(client rpc.Client) (string, error) {
	version := new(string)
	err := client.Call(Web3ClientVersionMethod, nil, version)
	if err != nil {
		return "", err
	}
	return *version, nil
}

// Wait for a transaction to be mined/confirmed
func AwaitTransaction(ctx context.Context, client rpc.Client, txHash string) (*Receipt, error) {
	for {
		tx, err := EthGetTransactionReceipt(client, txHash)
		if err != nil {
			return nil, fmt.Errorf("AwaitTransaction failed to get ethereum transaction: %w", err)
		}
		if tx.BlockNumber != "" {
			if tx.BlockHash == "" {
				return nil, fmt.Errorf("expected Blockhash to be non-empty when BlockNumber is non-empty (%s)",
					tx.BlockNumber)
			}
			// Transaction has been confirmed (is included in a block)
			return tx, nil
		}
		time.Sleep(awaitTransactionSleep)
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("AwaitTransaction interrupted: %w", ctx.Err())
		default:

		}
	}
}

func ethLogBound(bound *rpcevents.Bound) string {
	if bound == nil {
		return ""
	}
	switch bound.Type {
	case rpcevents.Bound_FIRST:
		return "earliest"
	case rpcevents.Bound_LATEST:
		return "latest"
	case rpcevents.Bound_ABSOLUTE:
		return web3.HexEncoder.Uint64(bound.Index)
	default:
		return ""
	}
}
