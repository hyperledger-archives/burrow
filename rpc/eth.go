package rpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/acm/balance"
	"github.com/hyperledger/burrow/acm/validator"
	bcm "github.com/hyperledger/burrow/bcm"
	bin "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/crypto"
	x "github.com/hyperledger/burrow/encoding/hex"
	"github.com/hyperledger/burrow/encoding/rlp"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/project"
	"github.com/hyperledger/burrow/rpc/web3"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	tmConfig "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/types"
)

const (
	chainID      = 1
	maxGasLimit  = 2<<52 - 1
	hexZero      = "0x0"
	hexZeroNonce = "0x0000000000000000"
	pending      = "null"
)

// EthService is a web3 provider
type EthService struct {
	accounts   acmstate.IterableStatsReader
	events     EventsReader
	blockchain bcm.BlockchainInfo
	validators validator.History
	nodeView   *tendermint.NodeView
	trans      *execution.Transactor
	keyClient  keys.KeyClient
	keyStore   *keys.KeyStore
	config     *tmConfig.Config
	logger     *logging.Logger
}

// NewEthService returns our web3 provider
func NewEthService(accounts acmstate.IterableStatsReader,
	events EventsReader, blockchain bcm.BlockchainInfo,
	validators validator.History, nodeView *tendermint.NodeView,
	trans *execution.Transactor, keyStore *keys.KeyStore,
	logger *logging.Logger) *EthService {

	keyClient := keys.NewLocalKeyClient(keyStore, logger)

	return &EthService{
		accounts,
		events,
		blockchain,
		validators,
		nodeView,
		trans,
		keyClient,
		keyStore,
		tmConfig.DefaultConfig(),
		logger,
	}
}

var _ web3.Service = &EthService{}

type EventsReader interface {
	TxsAtHeight(height uint64) ([]*exec.TxExecution, error)
	TxByHash(txHash []byte) (*exec.TxExecution, error)
}

var _ EventsReader = &state.State{}

// Web3ClientVersion returns the version of burrow
func (srv *EthService) Web3ClientVersion() (*web3.Web3ClientVersionResult, error) {
	return &web3.Web3ClientVersionResult{
		ClientVersion: project.FullVersion(),
	}, nil
}

// Web3Sha3 returns Keccak-256 (not the standardized SHA3-256) of the given data
func (srv *EthService) Web3Sha3(req *web3.Web3Sha3Params) (*web3.Web3Sha3Result, error) {
	data, err := x.DecodeToBytes(req.Data)
	if err != nil {
		return nil, err
	}

	return &web3.Web3Sha3Result{
		HashedData: x.EncodeBytes(crypto.Keccak256(data)),
	}, nil
}

// NetListening returns true if the peer is running
func (srv *EthService) NetListening() (*web3.NetListeningResult, error) {
	return &web3.NetListeningResult{
		IsNetListening: srv.nodeView.NodeInfo().GetListenAddress() != "",
	}, nil
}

// NetPeerCount returns the number of connected peers
func (srv *EthService) NetPeerCount() (*web3.NetPeerCountResult, error) {
	return &web3.NetPeerCountResult{
		NumConnectedPeers: x.EncodeNumber(uint64(srv.nodeView.Peers().Size())),
	}, nil
}

// NetVersion returns the hex encoding of the network id,
// this is typically a small int (where 1 == Ethereum mainnet)
func (srv *EthService) NetVersion() (*web3.NetVersionResult, error) {
	return &web3.NetVersionResult{
		ChainID: x.EncodeNumber(uint64(chainID)),
	}, nil
}

// EthProtocolVersion returns the version of tendermint
func (srv *EthService) EthProtocolVersion() (*web3.EthProtocolVersionResult, error) {
	return &web3.EthProtocolVersionResult{
		ProtocolVersion: srv.nodeView.NodeInfo().Version,
	}, nil
}

// EthChainId returns the chainID
func (srv *EthService) EthChainId() (*web3.EthChainIdResult, error) {
	return &web3.EthChainIdResult{
		ChainId: srv.blockchain.ChainID(),
	}, nil
}

// EthBlockNumber returns the latest height
func (srv *EthService) EthBlockNumber() (*web3.EthBlockNumberResult, error) {
	return &web3.EthBlockNumberResult{
		BlockNumber: x.EncodeNumber(srv.blockchain.LastBlockHeight()),
	}, nil
}

// EthCall executes a new message call immediately without creating a transaction
func (srv *EthService) EthCall(req *web3.EthCallParams) (*web3.EthCallResult, error) {
	var to, from crypto.Address
	var err error

	if addr := req.Transaction.To; addr != "" {
		to, err = x.DecodeToAddress(addr)
		if err != nil {
			return nil, err
		}
	}

	if addr := req.Transaction.From; addr != "" {
		from, err = x.DecodeToAddress(addr)
		if err != nil {
			return nil, err
		}
	}

	data, err := x.DecodeToBytes(req.Transaction.Data)
	if err != nil {
		return nil, err
	}

	txe, err := execution.CallSim(srv.accounts, srv.blockchain, from, to, data, srv.logger)
	if err != nil {
		return nil, err
	} else if txe.Exception != nil {
		return nil, txe.Exception.AsError()
	}

	var result string
	if r := txe.GetResult(); r != nil {
		result = x.EncodeBytes(r.GetReturn())
	}

	return &web3.EthCallResult{
		ReturnValue: result,
	}, nil
}

// EthGetBalance returns an accounts balance, or an error if it does not exist
func (srv *EthService) EthGetBalance(req *web3.EthGetBalanceParams) (*web3.EthGetBalanceResult, error) {
	addr, err := x.DecodeToAddress(req.Address)
	if err != nil {
		return nil, err
	}

	// TODO: read account state at height
	acc, err := srv.accounts.GetAccount(addr)
	if err != nil {
		return nil, err
	} else if acc == nil {
		return nil, fmt.Errorf("account not found at address %s", req.Address)
	}

	return &web3.EthGetBalanceResult{
		GetBalanceResult: x.EncodeBytes(balance.NativeToWei(acc.Balance).Bytes()),
	}, nil
}

// EthGetBlockByHash iterates through all headers to find a matching block height for a given hash
func (srv *EthService) EthGetBlockByHash(req *web3.EthGetBlockByHashParams) (*web3.EthGetBlockByHashResult, error) {
	height, err := srv.getBlockHeightByHash(req.BlockHash)
	if err != nil {
		return nil, err
	}

	block, err := srv.getBlockInfoAtHeight(height, req.IsTransactionsIncluded)
	if err != nil {
		return nil, err
	}

	return &web3.EthGetBlockByHashResult{
		GetBlockByHashResult: block,
	}, nil
}

// EthGetBlockByNumber returns block info at the given height
func (srv *EthService) EthGetBlockByNumber(req *web3.EthGetBlockByNumberParams) (*web3.EthGetBlockByNumberResult, error) {
	height, err := srv.getHeightByWordOrNumber(req.BlockNumber)
	if err != nil {
		return nil, err
	}

	block, err := srv.getBlockInfoAtHeight(height, req.IsTransactionsIncluded)
	if err != nil {
		return nil, err
	}

	return &web3.EthGetBlockByNumberResult{
		GetBlockByNumberResult: block,
	}, nil
}

// EthGetBlockTransactionCountByHash returns the number of transactions in a block matching a given hash
func (srv *EthService) EthGetBlockTransactionCountByHash(req *web3.EthGetBlockTransactionCountByHashParams) (*web3.EthGetBlockTransactionCountByHashResult, error) {
	height, err := srv.getBlockHeightByHash(req.BlockHash)
	if err != nil {
		return nil, err
	}

	block, err := srv.getBlockHeaderAtHeight(height)
	if err != nil {
		return nil, err
	}

	return &web3.EthGetBlockTransactionCountByHashResult{
		BlockTransactionCountByHash: x.EncodeNumber(uint64(block.NumTxs)),
	}, nil
}

// EthGetBlockTransactionCountByNumber returns the number of transactions in a block matching a given height
func (srv *EthService) EthGetBlockTransactionCountByNumber(req *web3.EthGetBlockTransactionCountByNumberParams) (*web3.EthGetBlockTransactionCountByNumberResult, error) {
	height, err := srv.getHeightByWordOrNumber(req.BlockNumber)
	if err != nil {
		return nil, err
	}

	block, err := srv.getBlockHeaderAtHeight(height)
	if err != nil {
		return nil, err
	}

	return &web3.EthGetBlockTransactionCountByNumberResult{
		BlockTransactionCountByHash: x.EncodeNumber(uint64(block.NumTxs)),
	}, nil
}

// EthGetCode returns the EVM bytecode at an address
func (srv *EthService) EthGetCode(req *web3.EthGetCodeParams) (*web3.EthGetCodeResult, error) {
	addr, err := x.DecodeToAddress(req.Address)
	if err != nil {
		return nil, err
	}

	acc, err := srv.accounts.GetAccount(addr)
	if err != nil {
		return nil, err
	} else if acc == nil {
		return nil, fmt.Errorf("account not found at address %s", req.Address)
	}

	return &web3.EthGetCodeResult{
		Bytes: x.EncodeBytes(acc.EVMCode),
	}, nil
}

func (srv *EthService) EthGetStorageAt(req *web3.EthGetStorageAtParams) (*web3.EthGetStorageAtResult, error) {
	// TODO
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthGetTransactionByBlockHashAndIndex(req *web3.EthGetTransactionByBlockHashAndIndexParams) (*web3.EthGetTransactionByBlockHashAndIndexResult, error) {
	height, err := srv.getBlockHeightByHash(req.BlockHash)
	if err != nil {
		return nil, err
	}

	head, err := srv.blockchain.GetBlockHeader(height)
	if err != nil {
		return nil, err
	}

	txes, err := srv.events.TxsAtHeight(height)
	if err != nil {
		return nil, err
	}

	index, err := x.DecodeToNumber(req.Index)
	if err != nil {
		return nil, err
	}

	for _, txe := range txes {
		if txe.GetIndex() == index {
			hash, tx, err := getHashAndCallTxFromExecution(txe)
			if err != nil {
				return nil, err
			}
			return &web3.EthGetTransactionByBlockHashAndIndexResult{
				TransactionResult: getTransaction(head, hash, tx),
			}, nil
		}
	}

	return nil, fmt.Errorf("tx not found at hash %s, index %d", req.BlockHash, index)
}

func (srv *EthService) EthGetTransactionByBlockNumberAndIndex(req *web3.EthGetTransactionByBlockNumberAndIndexParams) (*web3.EthGetTransactionByBlockNumberAndIndexResult, error) {
	height, err := srv.getHeightByWordOrNumber(req.BlockNumber)
	if err != nil {
		return nil, err
	}

	head, err := srv.blockchain.GetBlockHeader(height)
	if err != nil {
		return nil, err
	}

	txes, err := srv.events.TxsAtHeight(height)
	if err != nil {
		return nil, err
	}

	index, err := x.DecodeToNumber(req.Index)
	if err != nil {
		return nil, err
	}

	for _, txe := range txes {
		if txe.GetIndex() == index {
			hash, tx, err := getHashAndCallTxFromExecution(txe)
			if err != nil {
				return nil, err
			}
			return &web3.EthGetTransactionByBlockNumberAndIndexResult{
				TransactionResult: getTransaction(head, hash, tx),
			}, nil
		}
	}

	return nil, fmt.Errorf("tx not found at height %d, index %d", height, index)
}

// EthGetTransactionByHash finds a tx by the given hash
func (srv *EthService) EthGetTransactionByHash(req *web3.EthGetTransactionByHashParams) (*web3.EthGetTransactionByHashResult, error) {
	hash, err := x.DecodeToBytes(req.TransactionHash)
	if err != nil {
		return nil, err
	}

	txe, err := srv.events.TxByHash(hash)
	if err != nil {
		return nil, err
	}

	head, err := srv.blockchain.GetBlockHeader(txe.Height)
	if err != nil {
		return nil, err
	}

	hash, tx, err := getHashAndCallTxFromExecution(txe)
	if err != nil {
		return nil, err
	}

	return &web3.EthGetTransactionByHashResult{
		Transaction: getTransaction(head, hash, tx),
	}, nil
}

// EthGetTransactionCount returns the number of transactions sent from an address
func (srv *EthService) EthGetTransactionCount(req *web3.EthGetTransactionCountParams) (*web3.EthGetTransactionCountResult, error) {
	addr, err := x.DecodeToAddress(req.Address)
	if err != nil {
		return nil, err
	}

	// TODO: get tx count at height
	acc, err := srv.accounts.GetAccount(addr)
	if err != nil {
		return nil, err
	}

	// TODO: sequence may not always be accurate, is there a better way?
	return &web3.EthGetTransactionCountResult{
		NonceOrNull: x.EncodeNumber(acc.GetSequence()),
	}, nil
}

func getHashAndCallTxFromEnvelope(env *txs.Envelope) ([]byte, *payload.CallTx, error) {
	if env.Tx == nil {
		return nil, nil, fmt.Errorf("tx not found for %s", env.String())
	} else if tx, ok := env.Tx.Payload.(*payload.CallTx); ok {
		return env.Tx.Hash().Bytes(), tx, nil
	}
	return nil, nil, fmt.Errorf("tx not valid")
}

func getHashAndCallTxFromExecution(txe *exec.TxExecution) ([]byte, *payload.CallTx, error) {
	if txe.Envelope == nil {
		return nil, nil, fmt.Errorf("envelope not found for %s", txe.GetTxHash().String())
	}
	return getHashAndCallTxFromEnvelope(txe.Envelope)
}

// EthGetTransactionReceipt returns the receipt of a previously committed tx
func (srv *EthService) EthGetTransactionReceipt(req *web3.EthGetTransactionReceiptParams) (*web3.EthGetTransactionReceiptResult, error) {
	data, err := x.DecodeToBytes(req.TransactionHash)
	if err != nil {
		return nil, err
	}

	txe, err := srv.events.TxByHash(data)
	if err != nil {
		return nil, err
	} else if txe == nil {
		return nil, fmt.Errorf("tx with hash %s does not exist", req.TransactionHash)
	}

	hash, tx, err := getHashAndCallTxFromExecution(txe)
	if err != nil {
		return nil, err
	}

	block, err := srv.blockchain.GetBlockHeader(txe.Height)
	if err != nil {
		return nil, err
	}

	status := x.EncodeNumber(1)
	if err := txe.Exception.AsError(); err != nil {
		status = x.EncodeNumber(0)
	}

	result := &web3.EthGetTransactionReceiptResult{
		Receipt: web3.Receipt{
			Status:            status,
			TransactionIndex:  x.EncodeNumber(txe.GetIndex()),
			BlockNumber:       x.EncodeNumber(uint64(block.Height)),
			BlockHash:         x.EncodeBytes(block.Hash()),
			From:              x.EncodeBytes(tx.GetInput().Address.Bytes()),
			GasUsed:           x.EncodeNumber(txe.Result.GetGasUsed()),
			TransactionHash:   x.EncodeBytes(hash),
			CumulativeGasUsed: hexZero,
			LogsBloom:         hexZero,
			Logs:              []web3.Logs{},
		},
	}

	if txe.Receipt != nil {
		result.Receipt.ContractAddress = x.EncodeBytes(txe.Receipt.ContractAddress.Bytes())
		result.Receipt.To = pending
	} else if tx.Address != nil {
		result.Receipt.To = x.EncodeBytes(tx.Address.Bytes())
	}

	return result, nil
}

// EthHashrate returns the configured tendermint commit timeout
func (srv *EthService) EthHashrate() (*web3.EthHashrateResult, error) {
	return &web3.EthHashrateResult{
		HashesPerSecond: srv.config.Consensus.TimeoutCommit.String(),
	}, nil
}

// EthMining returns true if client is a validator
func (srv *EthService) EthMining() (*web3.EthMiningResult, error) {
	var isVal bool
	addr := srv.nodeView.ValidatorAddress()
	val := srv.validators.Validators(1)
	err := val.IterateValidators(func(id crypto.Addressable, _ *big.Int) error {
		if addr == id.GetAddress() {
			isVal = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &web3.EthMiningResult{
		Mining: isVal,
	}, nil
}

// EthPendingTransactions returns all txs in the mempool
func (srv *EthService) EthPendingTransactions() (*web3.EthPendingTransactionsResult, error) {
	pending := make([]web3.PendingTransactions, 0)
	envelopes, err := srv.nodeView.MempoolTransactions(-1)
	if err != nil {
		return nil, err
	}

	for _, env := range envelopes {
		hash, tx, err := getHashAndCallTxFromEnvelope(env)
		if err != nil {
			continue
		}
		pending = append(pending, web3.PendingTransactions{
			Transaction: getTransaction(nil, hash, tx),
		})
	}

	return &web3.EthPendingTransactionsResult{
		PendingTransactions: pending,
	}, nil
}

func (srv *EthService) EthEstimateGas(req *web3.EthEstimateGasParams) (*web3.EthEstimateGasResult, error) {
	// TODO
	return &web3.EthEstimateGasResult{
		GasUsed: hexZero,
	}, nil
}

func (srv *EthService) EthGasPrice() (*web3.EthGasPriceResult, error) {
	// TODO
	return &web3.EthGasPriceResult{
		GasPrice: hexZero,
	}, nil
}

type RawTx struct {
	Nonce    uint64 `json:"nonce"`
	GasPrice uint64 `json:"gasPrice"`
	GasLimit uint64 `json:"gasLimit"`
	To       []byte `json:"to"`
	Value    []byte `json:"value"`
	Data     []byte `json:"data"`

	V uint64 `json:"v"`
	R []byte `json:"r"`
	S []byte `json:"s"`
}

func (srv *EthService) EthGetRawTransactionByHash(req *web3.EthGetRawTransactionByHashParams) (*web3.EthGetRawTransactionByHashResult, error) {
	// TODO
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthGetRawTransactionByBlockHashAndIndex(req *web3.EthGetRawTransactionByBlockHashAndIndexParams) (*web3.EthGetRawTransactionByBlockHashAndIndexResult, error) {
	// TODO
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthGetRawTransactionByBlockNumberAndIndex(req *web3.EthGetRawTransactionByBlockNumberAndIndexParams) (*web3.EthGetRawTransactionByBlockNumberAndIndexResult, error) {
	// TODO
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthSendRawTransaction(req *web3.EthSendRawTransactionParams) (*web3.EthSendRawTransactionResult, error) {
	data, err := x.DecodeToBytes(req.SignedTransactionData)
	if err != nil {
		return nil, err
	}

	rawTx := new(RawTx)
	err = rlp.Decode(data, rawTx)
	if err != nil {
		return nil, err
	}

	net := uint64(chainID)
	enc, err := txs.RLPEncode(rawTx.Nonce, rawTx.GasPrice, rawTx.GasLimit, rawTx.To, rawTx.Value, rawTx.Data)
	if err != nil {
		return nil, err
	}

	sig := crypto.CompressedSignatureFromParams(rawTx.V-net-8-1, rawTx.R, rawTx.S)
	pub, err := crypto.PublicKeyFromSignature(sig, crypto.Keccak256(enc))
	if err != nil {
		return nil, err
	}
	from := pub.GetAddress()
	unc := crypto.UncompressedSignatureFromParams(rawTx.R, rawTx.S)
	signature, err := crypto.SignatureFromBytes(unc, crypto.CurveTypeSecp256k1)
	if err != nil {
		return nil, err
	}

	to, err := crypto.AddressFromBytes(rawTx.To)
	if err != nil {
		return nil, err
	}

	amount := balance.WeiToNative(rawTx.Value).Uint64()

	txEnv := &txs.Envelope{
		Signatories: []txs.Signatory{
			{
				Address:   &from,
				PublicKey: pub,
				Signature: signature,
			},
		},
		Enc: txs.Envelope_RLP,
		Tx: &txs.Tx{
			ChainID: srv.blockchain.ChainID(),
			Payload: &payload.CallTx{
				Input: &payload.TxInput{
					Address: from,
					Amount:  amount,
					// first tx sequence should be 1,
					// but metamask starts at 0
					Sequence: rawTx.Nonce + 1,
				},
				Address:  &to,
				GasLimit: rawTx.GasLimit,
				GasPrice: rawTx.GasPrice,
				Data:     rawTx.Data,
			},
		},
	}

	ctx := context.Background()
	txe, err := srv.trans.BroadcastTxSync(ctx, txEnv)
	if err != nil {
		return nil, err
	} else if txe.Exception != nil {
		return nil, txe.Exception.AsError()
	}

	return &web3.EthSendRawTransactionResult{
		TransactionHash: x.EncodeBytes(txe.GetTxHash().Bytes()),
	}, nil
}

// EthSyncing returns this nodes syncing status (i.e. whether it has caught up)
func (srv *EthService) EthSyncing() (*web3.EthSyncingResult, error) {
	// TODO: remaining sync fields
	return &web3.EthSyncingResult{
		Syncing: web3.SyncStatus{
			CurrentBlock: x.EncodeNumber(srv.blockchain.LastBlockHeight()),
		},
	}, nil
}

func (srv *EthService) getBlockHeightByHash(hash string) (uint64, error) {
	for i := uint64(1); i < srv.blockchain.LastBlockHeight(); i++ {
		head, err := srv.blockchain.GetBlockHeader(i)
		if err != nil {
			return 0, err
		} else if hexKeccak(head.Hash().Bytes()) == hash {
			return i, nil
		}
	}
	return 0, web3.ErrNotFound
}

func (srv *EthService) getBlockHeaderAtHeight(height uint64) (*types.Header, error) {
	return srv.blockchain.GetBlockHeader(height)
}

func hexKeccak(data []byte) string {
	return x.EncodeBytes(crypto.Keccak256(data))
}

func hexKeccakAddress(data []byte) string {
	addr := crypto.Keccak256(data)
	return x.EncodeBytes(addr[len(addr)-20:])
}

func (srv *EthService) getBlockInfoAtHeight(height uint64, includeTxs bool) (web3.Block, error) {
	doc := srv.blockchain.GenesisDoc()
	if height == 0 {
		// genesis
		return web3.Block{
			Transactions:    make([]web3.Transactions, 0),
			Uncles:          make([]string, 0),
			Nonce:           hexZeroNonce,
			Hash:            hexKeccak(doc.AppHash.Bytes()),
			ParentHash:      hexKeccak(doc.AppHash.Bytes()),
			ReceiptsRoot:    hexKeccak(doc.AppHash.Bytes()),
			StateRoot:       hexKeccak(doc.AppHash.Bytes()),
			Miner:           x.EncodeBytes(doc.Validators[0].Address.Bytes()),
			Timestamp:       x.EncodeNumber(uint64(doc.GenesisTime.Unix())),
			Number:          hexZero,
			Size:            hexZero,
			ExtraData:       hexZero,
			Difficulty:      hexZero,
			TotalDifficulty: hexZero,
			GasLimit:        hexZero,
			GasUsed:         hexZero,
		}, nil
	}
	block, err := srv.getBlockHeaderAtHeight(height)
	if err != nil {
		return web3.Block{}, err
	} else if block == nil {
		return web3.Block{}, fmt.Errorf("block at height %d does not exist", height)
	}

	transactions := make([]web3.Transactions, 0)
	if includeTxs {
		txes, err := srv.events.TxsAtHeight(height)
		if err != nil {
			return web3.Block{}, err
		}
		for _, txe := range txes {
			hash, tx, err := getHashAndCallTxFromExecution(txe)
			if err != nil {
				continue
			}
			transactions = append(transactions, web3.Transactions{
				getTransaction(block, hash, tx),
			})
		}
	}

	return web3.Block{
		Hash:             hexKeccak(block.Hash().Bytes()),
		ParentHash:       hexKeccak(block.Hash().Bytes()),
		TransactionsRoot: hexKeccak(block.Hash().Bytes()),
		StateRoot:        hexKeccak(block.Hash().Bytes()),
		ReceiptsRoot:     hexKeccak(block.Hash().Bytes()),
		Nonce:            hexZeroNonce,
		Size:             x.EncodeNumber(uint64(block.TotalTxs)),
		Number:           x.EncodeNumber(uint64(block.Height)),
		Miner:            x.EncodeBytes(block.ProposerAddress.Bytes()),
		Sha3Uncles:       hexZero,
		LogsBloom:        hexZero,
		ExtraData:        hexZero,
		Difficulty:       hexZero,
		TotalDifficulty:  hexZero,
		GasUsed:          hexZero,
		GasLimit:         x.EncodeNumber(maxGasLimit),
		Timestamp:        x.EncodeNumber(uint64(block.Time.Unix())),
		Transactions:     transactions,
		Uncles:           []string{},
	}, nil
}

func getTransaction(block *types.Header, hash []byte, tx *payload.CallTx) web3.Transaction {
	// TODO: sensible defaults for non-call
	transaction := web3.Transaction{
		V:        hexZero,
		R:        hexZero,
		S:        hexZero,
		From:     x.EncodeBytes(tx.Input.Address.Bytes()),
		Value:    x.EncodeNumber(tx.Input.Amount),
		Nonce:    x.EncodeNumber(tx.Input.Sequence),
		Gas:      x.EncodeNumber(tx.GasLimit),
		GasPrice: x.EncodeNumber(tx.GasPrice),
		Data:     x.EncodeBytes(tx.Data),
	}

	if block != nil {
		// may be pending
		transaction.BlockHash = hexKeccak(block.Hash().Bytes())
		transaction.Hash = x.EncodeBytes(hash)
		transaction.BlockNumber = x.EncodeNumber(uint64(block.Height))
		transaction.TransactionIndex = hexZero
	}

	if tx.Address != nil {
		transaction.To = x.EncodeBytes(tx.Address.Bytes())
	} else {
		transaction.To = pending
	}

	return transaction
}

func (srv *EthService) getHeightByWord(height string) (uint64, bool) {
	switch height {
	case "earliest":
		return 0, true
	case "latest", "pending":
		return srv.blockchain.LastBlockHeight(), true
		// TODO: pending state/transactions
	default:
		return 0, false
	}
}

func getHeightByNumber(height string) (uint64, error) {
	return x.DecodeToNumber(height)
}

func (srv *EthService) getHeightByWordOrNumber(i string) (uint64, error) {
	var err error
	height, ok := srv.getHeightByWord(i)
	if !ok {
		height, err = getHeightByNumber(i)
		if err != nil {
			return 0, err
		}
	}
	return height, nil
}

// EthSendTransaction constructs, signs and broadcasts a tx from the local node
// Note: https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1767.md#rationale
func (srv *EthService) EthSendTransaction(req *web3.EthSendTransactionParams) (*web3.EthSendTransactionResult, error) {
	tx := &payload.CallTx{
		Input: new(payload.TxInput),
	}

	var err error
	if from := req.Transaction.From; from != "" {
		tx.Input.Address, err = x.DecodeToAddress(from)
		if err != nil {
			return nil, fmt.Errorf("failed to parse from address: %v", err)
		}
	} else {
		return nil, fmt.Errorf("no from address specified")
	}

	if value := req.Transaction.Value; value != "" {
		tx.Input.Amount, err = strconv.ParseUint(value, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount: %v", err)
		}
	}

	acc, err := srv.accounts.GetAccount(tx.Input.Address)
	if err != nil {
		return nil, err
	} else if acc == nil {
		return nil, fmt.Errorf("account %s does not exist", tx.Input.Address.String())
	}

	tx.Input.Sequence = acc.Sequence + 1

	if to := req.Transaction.To; to != "" {
		addr, err := x.DecodeToAddress(to)
		if err != nil {
			return nil, fmt.Errorf("failed to parse to address: %v", err)
		}
		tx.Address = &addr
	}

	// gas provided for the transaction execution
	if gasLimit := req.Transaction.Gas; gasLimit != "" {
		tx.GasLimit, err = strconv.ParseUint(gasLimit, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse gasLimit: %v", err)
		}
	}

	if gasPrice := req.Transaction.GasPrice; gasPrice != "" {
		tx.GasPrice, err = strconv.ParseUint(gasPrice, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse gasPrice: %v", err)
		}
	}

	if data := req.Transaction.Data; data != "" {
		bs, err := x.DecodeToBytes(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse data: %v", err)
		}
		tx.Data = bin.HexBytes(bs)
	}

	txEnv := txs.Enclose(srv.blockchain.ChainID(), tx)

	signer, err := keys.AddressableSigner(srv.keyClient, tx.Input.Address)
	if err != nil {
		return nil, err
	}
	err = txEnv.Sign(signer)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	txe, err := srv.trans.BroadcastTxSync(ctx, txEnv)
	if err != nil {
		return nil, err
	} else if txe.Exception != nil {
		return nil, txe.Exception.AsError()
	}

	return &web3.EthSendTransactionResult{
		TransactionHash: x.EncodeBytes(txe.GetTxHash().Bytes()),
	}, nil
}

// EthAccounts returns all accounts signable from the local node
func (srv *EthService) EthAccounts() (*web3.EthAccountsResult, error) {
	addresses, err := srv.keyStore.GetAllAddresses()
	if err != nil {
		return nil, err
	}

	addrs := make([]string, 0, len(addresses))
	for _, addr := range addresses {
		data, err := hex.DecodeString(addr)
		if err != nil {
			return nil, fmt.Errorf("could not decode address %s", addr)
		}
		key, err := srv.keyStore.GetKey("", data)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve key for %s", addr)
		} else if key.CurveType != crypto.CurveTypeSecp256k1 {
			// we only want ethereum keys
			continue
		}
		// TODO: only return accounts that exist in current chain
		addrs = append(addrs, x.EncodeBytes(key.Address.Bytes()))
	}

	return &web3.EthAccountsResult{
		Addresses: addrs,
	}, nil
}

// EthSign: https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_sign
func (srv *EthService) EthSign(req *web3.EthSignParams) (*web3.EthSignResult, error) {
	addr, err := x.DecodeToBytes(req.Address)
	if err != nil {
		return nil, err
	}
	to, err := crypto.AddressFromBytes(addr)
	if err != nil {
		return nil, err
	}
	signer, err := keys.AddressableSigner(srv.keyClient, to)
	if err != nil {
		return nil, err
	}

	data, err := x.DecodeToBytes(req.Bytes)
	if err != nil {
		return nil, err
	}

	msg := append([]byte{0x19}, []byte("Ethereum Signed Message:\n")...)
	msg = append(msg, byte(len(data)))
	msg = append(msg, data...)

	sig, err := signer.Sign(crypto.Keccak256(msg))
	if err != nil {
		return nil, err
	}

	return &web3.EthSignResult{
		Signature: x.EncodeBytes(sig.RawBytes()),
	}, nil
}

// N / A

func (srv *EthService) EthUninstallFilter(*web3.EthUninstallFilterParams) (*web3.EthUninstallFilterResult, error) {
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthSubmitHashrate(req *web3.EthSubmitHashrateParams) (*web3.EthSubmitHashrateResult, error) {
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthSubmitWork(*web3.EthSubmitWorkParams) (*web3.EthSubmitWorkResult, error) {
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthNewBlockFilter() (*web3.EthNewBlockFilterResult, error) {
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthNewFilter(req *web3.EthNewFilterParams) (*web3.EthNewFilterResult, error) {
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthNewPendingTransactionFilter() (*web3.EthNewPendingTransactionFilterResult, error) {
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthGetUncleByBlockHashAndIndex(req *web3.EthGetUncleByBlockHashAndIndexParams) (*web3.EthGetUncleByBlockHashAndIndexResult, error) {
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthGetUncleByBlockNumberAndIndex(req *web3.EthGetUncleByBlockNumberAndIndexParams) (*web3.EthGetUncleByBlockNumberAndIndexResult, error) {
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthGetUncleCountByBlockHash(req *web3.EthGetUncleCountByBlockHashParams) (*web3.EthGetUncleCountByBlockHashResult, error) {
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthGetUncleCountByBlockNumber(req *web3.EthGetUncleCountByBlockNumberParams) (*web3.EthGetUncleCountByBlockNumberResult, error) {
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthGetProof(req *web3.EthGetProofParams) (*web3.EthGetProofResult, error) {
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthGetWork() (*web3.EthGetWorkResult, error) {
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthGetFilterChanges(req *web3.EthGetFilterChangesParams) (*web3.EthGetFilterChangesResult, error) {
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthGetFilterLogs(req *web3.EthGetFilterLogsParams) (*web3.EthGetFilterLogsResult, error) {
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthCoinbase() (*web3.EthCoinbaseResult, error) {
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthGetLogs(req *web3.EthGetLogsParams) (*web3.EthGetLogsResult, error) {
	return nil, web3.ErrNotFound
}
