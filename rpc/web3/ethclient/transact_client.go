package ethclient

import (
	"context"
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/encoding/web3hex"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"google.golang.org/grpc"
)

const BasicGasLimit = 21000

// Provides a partial implementation of the GRPC-generated TransactClient suitable for testing Vent on Ethereum
type TransactClient struct {
	client   ethClient
	chainID  string
	accounts []acm.AddressableSigner
}

type ethClient interface {
	AwaitTransaction(ctx context.Context, txHash string) (*Receipt, error)
	SendTransaction(tx *EthSendTransactionParam) (string, error)
	SendRawTransaction(txHex string) (string, error)
	GetTransactionCount(address crypto.Address) (string, error)
	NetVersion() (string, error)
	GasPrice() (string, error)
}

func NewTransactClient(client ethClient) *TransactClient {
	return &TransactClient{
		client: client,
	}
}

func (cli *TransactClient) WithAccounts(signers ...acm.AddressableSigner) *TransactClient {
	return &TransactClient{
		client:   cli.client,
		accounts: append(cli.accounts, signers...),
	}
}

func (cli *TransactClient) CallTxSync(ctx context.Context, tx *payload.CallTx,
	opts ...grpc.CallOption) (*exec.TxExecution, error) {

	var signer acm.AddressableSigner

	for _, sa := range cli.accounts {
		if sa.GetAddress() == tx.Input.Address {
			signer = sa
			break
		}
	}

	// Only set nonce for tx we sign, otherwise let server do it
	err := cli.completeTx(tx, signer != nil)
	if err != nil {
		return nil, fmt.Errorf("could not set values on transaction")
	}

	var txHash string
	if signer == nil {
		txHash, err = cli.SendTransaction(tx)
	} else {
		txHash, err = cli.SendRawTransaction(tx, signer)
	}
	if err != nil {
		return nil, fmt.Errorf("could not send ethereum transaction: %w", err)
	}

	fmt.Printf("Waiting for tranasaction %s to be confirmed...\n", txHash)
	receipt, err := cli.client.AwaitTransaction(ctx, txHash)
	if err != nil {
		return nil, err
	}

	d := new(web3hex.Decoder)

	header := &exec.TxHeader{
		TxType: payload.TypeCall,
		TxHash: d.Bytes(receipt.TransactionHash),
		Height: d.Uint64(receipt.BlockNumber),
		Index:  d.Uint64(receipt.TransactionIndex),
	}

	// Attempt to provide sufficient return values to satisfy Vent's needs.
	return &exec.TxExecution{
		TxHeader: header,
		Receipt: &txs.Receipt{
			TxType:          header.TxType,
			TxHash:          header.TxHash,
			CreatesContract: receipt.ContractAddress != "",
			ContractAddress: d.Address(receipt.ContractAddress),
		},
	}, d.Err()
}

func (cli *TransactClient) SendTransaction(tx *payload.CallTx) (string, error) {
	var to string
	if tx.Address != nil {
		to = web3hex.Encoder.Address(*tx.Address)
	}

	var nonce string
	if tx.Input.Sequence != 0 {
		nonce = web3hex.Encoder.Uint64OmitEmpty(tx.Input.Sequence)
	}

	param := &EthSendTransactionParam{
		From:     web3hex.Encoder.Address(tx.Input.Address),
		To:       to,
		Gas:      web3hex.Encoder.Uint64OmitEmpty(tx.GasLimit),
		GasPrice: web3hex.Encoder.Uint64OmitEmpty(tx.GasPrice),
		Value:    web3hex.Encoder.Uint64OmitEmpty(tx.Input.Amount),
		Data:     web3hex.Encoder.BytesTrim(tx.Data),
		Nonce:    nonce,
	}

	return cli.client.SendTransaction(param)
}

func (cli *TransactClient) SendRawTransaction(tx *payload.CallTx, signer acm.AddressableSigner) (string, error) {
	chainID, err := cli.GetChainID()
	if err != nil {
		return "", err
	}
	txEnv := txs.Enclose(chainID, tx)

	txEnv.Encoding = txs.Envelope_RLP

	err = txEnv.Sign(signer)
	if err != nil {
		return "", fmt.Errorf("could not sign Ethereum transaction: %w", err)
	}

	rawTx, err := txs.EthRawTxFromEnvelope(txEnv)
	if err != nil {
		return "", fmt.Errorf("could not generate Ethereum raw transaction: %w", err)
	}

	bs, err := rawTx.Marshal()
	if err != nil {
		return "", fmt.Errorf("could not marshal Ethereum raw transaction: %w", err)
	}

	return cli.client.SendRawTransaction(web3hex.Encoder.BytesTrim(bs))
}

func (cli *TransactClient) GetChainID() (string, error) {
	if cli.chainID == "" {
		var err error
		cli.chainID, err = cli.client.NetVersion()
		if err != nil {
			return "", fmt.Errorf("TransactClient could not get ChainID: %w", err)
		}
	}
	return cli.chainID, nil
}

func (cli *TransactClient) GetGasPrice() (uint64, error) {
	gasPrice, err := cli.client.GasPrice()
	if err != nil {
		return 0, fmt.Errorf("could not get gas price: %w", err)
	}
	d := new(web3hex.Decoder)
	return d.Uint64(gasPrice), d.Err()
}

func (cli *TransactClient) GetTransactionCount(address crypto.Address) (uint64, error) {
	count, err := cli.client.GetTransactionCount(address)
	if err != nil {
		return 0, fmt.Errorf("could not get transaction acount for address %s: %w", address, err)
	}
	d := new(web3hex.Decoder)
	return d.Uint64(count), d.Err()
}

func (cli *TransactClient) completeTx(tx *payload.CallTx, setNonce bool) error {
	if tx.GasLimit == 0 {
		tx.GasLimit = BasicGasLimit
	}
	var err error
	if tx.GasPrice == 0 {
		tx.GasPrice, err = cli.GetGasPrice()
		if err != nil {
			return err
		}
	}
	if setNonce && tx.Input.Sequence == 0 {
		tx.Input.Sequence, err = cli.GetTransactionCount(tx.Input.Address)
		if err != nil {
			return err
		}
	}
	return nil
}
