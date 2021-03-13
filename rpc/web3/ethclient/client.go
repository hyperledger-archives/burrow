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
	EthGetBlockByNumberMethod      = "eth_getBlockByNumber"
	EthGetTransactionByHashMethod  = "eth_getTransactionByHash"
	EthGetTransactionReceiptMethod = "eth_getTransactionReceipt"
	EthGasPriceMethod              = "eth_gasPrice"
	NetVersionMethod               = "net_version"
	Web3ClientVersionMethod        = "web3_clientVersion"
)

// Adjust the polling frequency of AwaitTransaction
const awaitTransactionSleep = 200 * time.Millisecond

type EthClient struct {
	rpc.Client
}

func NewEthClient(cli rpc.Client) *EthClient {
	return &EthClient{Client: cli}
}

func (c *EthClient) SendTransaction(tx *EthSendTransactionParam) (string, error) {
	hash := new(string)
	err := c.Call(EthSendTransactionMethod, []*EthSendTransactionParam{tx}, &hash)
	if err != nil {
		return "", err
	}
	return *hash, nil
}

func (c *EthClient) SendRawTransaction(txHex string) (string, error) {
	hash := new(string)
	err := c.Call(EthSendRawTransactionMethod, []string{txHex}, &hash)
	if err != nil {
		return "", err
	}
	return *hash, nil
}

func (c *EthClient) GetTransactionCount(address crypto.Address) (string, error) {
	var count string
	err := c.Call(EthGetTransactionCountMethod, []string{web3.HexEncoder.Address(address), "latest"}, &count)
	if err != nil {
		return "", err
	}
	return count, nil
}

func (c *EthClient) GetLogs(filter *Filter) ([]*EthLog, error) {
	var logs []*EthLog
	err := c.Call(EthGetLogsMethod, []*EthFilter{filter.EthFilter()}, &logs)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (c *EthClient) Accounts() ([]string, error) {
	var accounts []string
	err := c.Call(EthAccountsMethod, nil, &accounts)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (c *EthClient) GetBlockByNumber(height string) (*Block, error) {
	block := new(Block)
	err := c.Call(EthGetBlockByNumberMethod, []interface{}{height, false}, block)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (c *EthClient) GetTransactionByHash(txHash string) (*web3.Transaction, error) {
	tx := new(web3.Transaction)
	err := c.Call(EthGetTransactionByHashMethod, []string{txHash}, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (c *EthClient) GetTransactionReceipt(txHash string) (*Receipt, error) {
	tx := new(Receipt)
	err := c.Call(EthGetTransactionReceiptMethod, []string{txHash}, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (c *EthClient) Syncing() (bool, error) {
	syncing := new(bool)
	err := c.Call(EthSyncingMethod, nil, syncing)
	if err != nil {
		return false, err
	}
	return *syncing, nil
}

func (c *EthClient) BlockNumber() (uint64, error) {
	latestBlock := new(string)
	err := c.Call(EthBlockNumberMethod, nil, latestBlock)
	if err != nil {
		return 0, err
	}
	d := new(web3.HexDecoder)
	return d.Uint64(*latestBlock), d.Err()
}

func (c *EthClient) GasPrice() (string, error) {
	gasPrice := new(string)
	err := c.Call(EthGasPriceMethod, nil, gasPrice)
	if err != nil {
		return "", err
	}
	return *gasPrice, nil
}

// AKA ChainID
func (c *EthClient) NetVersion() (string, error) {
	version := new(string)
	err := c.Call(NetVersionMethod, nil, version)
	if err != nil {
		return "", err
	}
	return *version, nil
}

func (c *EthClient) Web3ClientVersion() (string, error) {
	version := new(string)
	err := c.Call(Web3ClientVersionMethod, nil, version)
	if err != nil {
		return "", err
	}
	return *version, nil
}

// Wait for a transaction to be mined/confirmed
func (c *EthClient) AwaitTransaction(ctx context.Context, txHash string) (*Receipt, error) {
	for {
		tx, err := c.GetTransactionReceipt(txHash)
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

func logBound(bound *rpcevents.Bound) string {
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
