package web3

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"

	"github.com/hyperledger/burrow/encoding/rlp"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/acm/balance"
	"github.com/hyperledger/burrow/acm/validator"
	bcm "github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/project"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	tmConfig "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/types"
)

const (
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
	keyStore   *keys.FilesystemKeyStore
	config     *tmConfig.Config
	chainID    *big.Int
	logger     *logging.Logger
}

// NewEthService returns our web3 provider
func NewEthService(
	accounts acmstate.IterableStatsReader,
	events EventsReader,
	blockchain bcm.BlockchainInfo,
	validators validator.History,
	nodeView *tendermint.NodeView,
	trans *execution.Transactor,
	keyStore *keys.FilesystemKeyStore,
	logger *logging.Logger,
) *EthService {

	keyClient := keys.NewLocalKeyClient(keyStore, logger)

	return &EthService{
		accounts:   accounts,
		events:     events,
		blockchain: blockchain,
		validators: validators,
		nodeView:   nodeView,
		trans:      trans,
		keyClient:  keyClient,
		keyStore:   keyStore,
		config:     tmConfig.DefaultConfig(),
		// Ethereum expects ChainID to be an integer value
		chainID: crypto.GetEthChainID(blockchain.ChainID()),
		logger:  logger,
	}
}

var _ Service = &EthService{}

type EventsReader interface {
	TxsAtHeight(height uint64) ([]*exec.TxExecution, error)
	TxByHash(txHash []byte) (*exec.TxExecution, error)
}

var _ EventsReader = &state.State{}

// Web3ClientVersion returns the version of burrow
func (srv *EthService) Web3ClientVersion() (*Web3ClientVersionResult, error) {
	return &Web3ClientVersionResult{
		ClientVersion: project.FullVersion(),
	}, nil
}

// Web3Sha3 returns Keccak-256 (not the standardized SHA3-256) of the given data
func (srv *EthService) Web3Sha3(req *Web3Sha3Params) (*Web3Sha3Result, error) {
	d := new(HexDecoder)
	return &Web3Sha3Result{
		HashedData: HexEncoder.Bytes(crypto.Keccak256(d.Bytes(req.Data))),
	}, d.Err()
}

// NetListening returns true if the peer is running
func (srv *EthService) NetListening() (*NetListeningResult, error) {
	return &NetListeningResult{
		IsNetListening: srv.nodeView.NodeInfo().GetListenAddress() != "",
	}, nil
}

// NetPeerCount returns the number of connected peers
func (srv *EthService) NetPeerCount() (*NetPeerCountResult, error) {
	s := HexEncoder.Uint64(uint64(srv.nodeView.Peers().Size()))
	return &NetPeerCountResult{
		s,
	}, nil
}

// NetVersion returns the hex encoding of the network id,
// this is typically a small int (where 1 == Ethereum mainnet)
func (srv *EthService) NetVersion() (*NetVersionResult, error) {
	return &NetVersionResult{
		ChainID: HexEncoder.BigInt(srv.chainID),
	}, nil
}

// EthProtocolVersion returns the version of tendermint
func (srv *EthService) EthProtocolVersion() (*EthProtocolVersionResult, error) {
	return &EthProtocolVersionResult{
		ProtocolVersion: srv.nodeView.NodeInfo().Version,
	}, nil
}

// EthChainId returns the chainID
func (srv *EthService) EthChainId() (*EthChainIdResult, error) {
	return &EthChainIdResult{
		ChainId: HexEncoder.BigInt(srv.chainID),
	}, nil
}

// EthBlockNumber returns the latest height
func (srv *EthService) EthBlockNumber() (*EthBlockNumberResult, error) {
	return &EthBlockNumberResult{
		BlockNumber: HexEncoder.Uint64(srv.blockchain.LastBlockHeight()),
	}, nil
}

// EthCall executes a new message call immediately without creating a transaction
func (srv *EthService) EthCall(req *EthCallParams) (*EthCallResult, error) {
	d := new(HexDecoder)

	from := d.Address(req.Transaction.From)
	to := d.Address(req.Transaction.To)
	data := d.Bytes(req.Transaction.Data)

	if d.Err() != nil {
		return nil, d.Err()
	}
	txe, err := execution.CallSim(srv.accounts, srv.blockchain, from, to, data, srv.logger)
	if err != nil {
		return nil, err
	} else if txe.Exception != nil {
		return nil, txe.Exception.AsError()
	}

	var result string
	if r := txe.GetResult(); r != nil {
		result = HexEncoder.Bytes(r.GetReturn())
	}

	return &EthCallResult{
		ReturnValue: result,
	}, nil
}

// EthGetBalance returns an accounts balance, or an error if it does not exist
func (srv *EthService) EthGetBalance(req *EthGetBalanceParams) (*EthGetBalanceResult, error) {
	d := new(HexDecoder)
	addr := d.Address(req.Address)
	if d.Err() != nil {
		return nil, d.Err()
	}

	// TODO: read account state at height
	acc, err := srv.accounts.GetAccount(addr)
	if err != nil {
		return nil, err
	} else if acc == nil {
		return nil, fmt.Errorf("account not found at address %s", req.Address)
	}

	return &EthGetBalanceResult{
		GetBalanceResult: HexEncoder.Bytes(balance.NativeToWei(acc.Balance).Bytes()),
	}, nil
}

// EthGetBlockByHash iterates through all headers to find a matching block height for a given hash
func (srv *EthService) EthGetBlockByHash(req *EthGetBlockByHashParams) (*EthGetBlockByHashResult, error) {
	height, err := srv.getBlockHeightByHash(req.BlockHash)
	if err != nil {
		return nil, err
	}

	block, err := srv.getBlockInfoAtHeight(height, req.IsTransactionsIncluded)
	if err != nil {
		return nil, err
	}

	return &EthGetBlockByHashResult{
		GetBlockByHashResult: block,
	}, nil
}

// EthGetBlockByNumber returns block info at the given height
func (srv *EthService) EthGetBlockByNumber(req *EthGetBlockByNumberParams) (*EthGetBlockByNumberResult, error) {
	height, err := srv.getHeightByWordOrNumber(req.BlockNumber)
	if err != nil {
		return nil, err
	}

	block, err := srv.getBlockInfoAtHeight(height, req.IsTransactionsIncluded)
	if err != nil {
		return nil, err
	}

	return &EthGetBlockByNumberResult{
		GetBlockByNumberResult: block,
	}, nil
}

// EthGetBlockTransactionCountByHash returns the number of transactions in a block matching a given hash
func (srv *EthService) EthGetBlockTransactionCountByHash(req *EthGetBlockTransactionCountByHashParams) (*EthGetBlockTransactionCountByHashResult, error) {
	height, err := srv.getBlockHeightByHash(req.BlockHash)
	if err != nil {
		return nil, err
	}

	numTxs, err := srv.blockchain.GetNumTxs(height)
	if err != nil {
		return nil, err
	}

	return &EthGetBlockTransactionCountByHashResult{
		BlockTransactionCountByHash: HexEncoder.Uint64(uint64(numTxs)),
	}, nil
}

// EthGetBlockTransactionCountByNumber returns the number of transactions in a block matching a given height
func (srv *EthService) EthGetBlockTransactionCountByNumber(req *EthGetBlockTransactionCountByNumberParams) (*EthGetBlockTransactionCountByNumberResult, error) {
	height, err := srv.getHeightByWordOrNumber(req.BlockNumber)
	if err != nil {
		return nil, err
	}

	numTxs, err := srv.blockchain.GetNumTxs(height)
	if err != nil {
		return nil, err
	}

	return &EthGetBlockTransactionCountByNumberResult{
		BlockTransactionCountByHash: HexEncoder.Uint64(uint64(numTxs)),
	}, nil
}

// EthGetCode returns the EVM bytecode at an address
func (srv *EthService) EthGetCode(req *EthGetCodeParams) (*EthGetCodeResult, error) {
	d := new(HexDecoder)
	addr := d.Address(req.Address)
	if d.Err() != nil {
		return nil, d.Err()
	}

	acc, err := srv.accounts.GetAccount(addr)
	if err != nil {
		return nil, err
	} else if acc == nil {
		return nil, fmt.Errorf("account not found at address %s", req.Address)
	}

	return &EthGetCodeResult{
		Bytes: HexEncoder.Bytes(acc.EVMCode),
	}, nil
}

func (srv *EthService) EthGetStorageAt(req *EthGetStorageAtParams) (*EthGetStorageAtResult, error) {
	// TODO
	return nil, ErrNotFound
}

func (srv *EthService) EthGetTransactionByBlockHashAndIndex(req *EthGetTransactionByBlockHashAndIndexParams) (*EthGetTransactionByBlockHashAndIndexResult, error) {
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

	d := new(HexDecoder)

	index := d.Uint64(req.Index)

	if d.Err() != nil {
		return nil, d.Err()
	}

	for _, txe := range txes {
		if txe.GetIndex() == index {
			hash, tx, err := getHashAndCallTxFromExecution(txe)
			if err != nil {
				return nil, err
			}
			return &EthGetTransactionByBlockHashAndIndexResult{
				TransactionResult: getTransaction(head, hash, tx),
			}, nil
		}
	}

	return nil, fmt.Errorf("tx not found at hash %s, index %d", req.BlockHash, index)
}

func (srv *EthService) EthGetTransactionByBlockNumberAndIndex(req *EthGetTransactionByBlockNumberAndIndexParams) (*EthGetTransactionByBlockNumberAndIndexResult, error) {
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
	d := new(HexDecoder)
	index := d.Uint64(req.Index)
	if d.Err() != nil {
		return nil, d.Err()
	}

	for _, txe := range txes {
		if txe.GetIndex() == index {
			hash, tx, err := getHashAndCallTxFromExecution(txe)
			if err != nil {
				return nil, err
			}
			return &EthGetTransactionByBlockNumberAndIndexResult{
				TransactionResult: getTransaction(head, hash, tx),
			}, nil
		}
	}

	return nil, fmt.Errorf("tx not found at height %d, index %d", height, index)
}

// EthGetTransactionByHash finds a tx by the given hash
func (srv *EthService) EthGetTransactionByHash(req *EthGetTransactionByHashParams) (*EthGetTransactionByHashResult, error) {
	d := new(HexDecoder)

	hash := d.Bytes(req.TransactionHash)
	if d.Err() != nil {
		return nil, d.Err()
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

	return &EthGetTransactionByHashResult{
		Transaction: getTransaction(head, hash, tx),
	}, nil
}

// EthGetTransactionCount returns the number of transactions sent from an address
func (srv *EthService) EthGetTransactionCount(req *EthGetTransactionCountParams) (*EthGetTransactionCountResult, error) {
	d := new(HexDecoder)
	addr := d.Address(req.Address)
	if d.Err() != nil {
		return nil, d.Err()
	}

	// TODO: get tx count at height
	acc, err := srv.accounts.GetAccount(addr)
	if err != nil {
		return nil, err
	}

	// TODO: sequence may not always be accurate, is there a better way?
	return &EthGetTransactionCountResult{
		NonceOrNull: HexEncoder.Uint64(acc.GetSequence()),
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
func (srv *EthService) EthGetTransactionReceipt(req *EthGetTransactionReceiptParams) (*EthGetTransactionReceiptResult, error) {
	d := new(HexDecoder)

	data := d.Bytes(req.TransactionHash)
	if d.Err() != nil {
		return nil, d.Err()
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

	status := HexEncoder.Uint64(1)
	if err := txe.Exception.AsError(); err != nil {
		status = HexEncoder.Uint64(0)
	}

	result := &EthGetTransactionReceiptResult{
		Receipt: Receipt{
			Status:            status,
			TransactionIndex:  HexEncoder.Uint64(txe.GetIndex()),
			BlockNumber:       HexEncoder.Uint64(uint64(block.Height)),
			BlockHash:         HexEncoder.Bytes(block.Hash()),
			From:              HexEncoder.Bytes(tx.GetInput().Address.Bytes()),
			GasUsed:           HexEncoder.Uint64(txe.Result.GetGasUsed()),
			TransactionHash:   HexEncoder.Bytes(hash),
			CumulativeGasUsed: hexZero,
			LogsBloom:         hexZero,
			Logs:              []Logs{},
		},
	}

	if txe.Receipt != nil {
		result.Receipt.ContractAddress = HexEncoder.Bytes(txe.Receipt.ContractAddress.Bytes())
		result.Receipt.To = pending
	} else if tx.Address != nil {
		result.Receipt.To = HexEncoder.Bytes(tx.Address.Bytes())
	}

	return result, nil
}

// EthHashrate returns the configured tendermint commit timeout
func (srv *EthService) EthHashrate() (*EthHashrateResult, error) {
	return &EthHashrateResult{
		HashesPerSecond: srv.config.Consensus.TimeoutCommit.String(),
	}, nil
}

// EthMining returns true if client is a validator
func (srv *EthService) EthMining() (*EthMiningResult, error) {
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
	return &EthMiningResult{
		Mining: isVal,
	}, nil
}

// EthPendingTransactions returns all txs in the mempool
func (srv *EthService) EthPendingTransactions() (*EthPendingTransactionsResult, error) {
	pending := make([]PendingTransactions, 0)
	envelopes, err := srv.nodeView.MempoolTransactions(-1)
	if err != nil {
		return nil, err
	}

	for _, env := range envelopes {
		hash, tx, err := getHashAndCallTxFromEnvelope(env)
		if err != nil {
			continue
		}
		pending = append(pending, PendingTransactions{
			Transaction: getTransaction(nil, hash, tx),
		})
	}

	return &EthPendingTransactionsResult{
		PendingTransactions: pending,
	}, nil
}

func (srv *EthService) EthEstimateGas(req *EthEstimateGasParams) (*EthEstimateGasResult, error) {
	// TODO
	return &EthEstimateGasResult{
		GasUsed: hexZero,
	}, nil
}

func (srv *EthService) EthGasPrice() (*EthGasPriceResult, error) {
	// TODO
	return &EthGasPriceResult{
		GasPrice: hexZero,
	}, nil
}

func (srv *EthService) EthGetRawTransactionByHash(req *EthGetRawTransactionByHashParams) (*EthGetRawTransactionByHashResult, error) {
	// TODO
	return nil, ErrNotFound
}

func (srv *EthService) EthGetRawTransactionByBlockHashAndIndex(req *EthGetRawTransactionByBlockHashAndIndexParams) (*EthGetRawTransactionByBlockHashAndIndexResult, error) {
	// TODO
	return nil, ErrNotFound
}

func (srv *EthService) EthGetRawTransactionByBlockNumberAndIndex(req *EthGetRawTransactionByBlockNumberAndIndexParams) (*EthGetRawTransactionByBlockNumberAndIndexResult, error) {
	// TODO
	return nil, ErrNotFound
}

func (srv *EthService) EthSendRawTransaction(req *EthSendRawTransactionParams) (*EthSendRawTransactionResult, error) {
	d := new(HexDecoder)

	data := d.Bytes(req.SignedTransactionData)

	if d.Err() != nil {
		return nil, d.Err()
	}

	rawTx := txs.NewEthRawTx(srv.chainID)
	err := rlp.Decode(data, rawTx)
	if err != nil {
		return nil, err
	}

	publicKey, signature, err := rawTx.RecoverPublicKey()
	if err != nil {
		return nil, err
	}

	from := publicKey.GetAddress()

	to, err := crypto.AddressFromBytes(rawTx.To)
	if err != nil {
		return nil, err
	}

	amount := balance.WeiToNative(rawTx.Amount).Uint64()

	txEnv := &txs.Envelope{
		Signatories: []txs.Signatory{
			{
				Address:   &from,
				PublicKey: publicKey,
				Signature: signature,
			},
		},
		Encoding: txs.Envelope_RLP,
		Tx: &txs.Tx{
			ChainID: srv.blockchain.ChainID(),
			Payload: &payload.CallTx{
				Input: &payload.TxInput{
					Address: from,
					Amount:  amount,
					// first tx sequence should be 1,
					// but metamask starts at 0
					Sequence: rawTx.Sequence + 1,
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

	return &EthSendRawTransactionResult{
		TransactionHash: HexEncoder.Bytes(txe.GetTxHash().Bytes()),
	}, nil
}

// EthSyncing returns this nodes syncing status (i.e. whether it has caught up)
func (srv *EthService) EthSyncing() (*EthSyncingResult, error) {
	// TODO: remaining sync fields
	return &EthSyncingResult{
		Syncing: SyncStatus{
			CurrentBlock: HexEncoder.Uint64(srv.blockchain.LastBlockHeight()),
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
	return 0, ErrNotFound
}

func (srv *EthService) getBlockHeaderAtHeight(height uint64) (*types.Header, error) {
	return srv.blockchain.GetBlockHeader(height)
}

func hexKeccak(data []byte) string {
	return HexEncoder.Bytes(crypto.Keccak256(data))
}

func hexKeccakAddress(data []byte) string {
	addr := crypto.Keccak256(data)
	return HexEncoder.Bytes(addr[len(addr)-20:])
}

func (srv *EthService) getBlockInfoAtHeight(height uint64, includeTxs bool) (Block, error) {
	doc := srv.blockchain.GenesisDoc()
	if height == 0 {
		// genesis
		return Block{
			Transactions:    make([]Transactions, 0),
			Uncles:          make([]string, 0),
			Nonce:           hexZeroNonce,
			Hash:            hexKeccak(doc.AppHash.Bytes()),
			ParentHash:      hexKeccak(doc.AppHash.Bytes()),
			ReceiptsRoot:    hexKeccak(doc.AppHash.Bytes()),
			StateRoot:       hexKeccak(doc.AppHash.Bytes()),
			Miner:           HexEncoder.Bytes(doc.Validators[0].Address.Bytes()),
			Timestamp:       HexEncoder.Uint64(uint64(doc.GenesisTime.Unix())),
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
		return Block{}, err
	} else if block == nil {
		return Block{}, fmt.Errorf("block at height %d does not exist", height)
	}

	numTxs, err := srv.blockchain.GetNumTxs(height)
	if err != nil {
		return Block{}, err
	}

	transactions := make([]Transactions, 0)
	if includeTxs {
		txes, err := srv.events.TxsAtHeight(height)
		if err != nil {
			return Block{}, err
		}
		for _, txe := range txes {
			hash, tx, err := getHashAndCallTxFromExecution(txe)
			if err != nil {
				continue
			}
			transactions = append(transactions, Transactions{
				getTransaction(block, hash, tx),
			})
		}
	}

	return Block{
		Hash:             hexKeccak(block.Hash().Bytes()),
		ParentHash:       hexKeccak(block.Hash().Bytes()),
		TransactionsRoot: hexKeccak(block.Hash().Bytes()),
		StateRoot:        hexKeccak(block.Hash().Bytes()),
		ReceiptsRoot:     hexKeccak(block.Hash().Bytes()),
		Nonce:            hexZeroNonce,
		Size:             HexEncoder.Uint64(uint64(numTxs)),
		Number:           HexEncoder.Uint64(uint64(block.Height)),
		Miner:            HexEncoder.Bytes(block.ProposerAddress.Bytes()),
		Sha3Uncles:       hexZero,
		LogsBloom:        hexZero,
		ExtraData:        hexZero,
		Difficulty:       hexZero,
		TotalDifficulty:  hexZero,
		GasUsed:          hexZero,
		GasLimit:         HexEncoder.Uint64(maxGasLimit),
		Timestamp:        HexEncoder.Uint64(uint64(block.Time.Unix())),
		Transactions:     transactions,
		Uncles:           []string{},
	}, nil
}

func getTransaction(block *types.Header, hash []byte, tx *payload.CallTx) Transaction {
	// TODO: sensible defaults for non-call
	transaction := Transaction{
		V:        hexZero,
		R:        hexZero,
		S:        hexZero,
		From:     HexEncoder.Bytes(tx.Input.Address.Bytes()),
		Value:    HexEncoder.Uint64(tx.Input.Amount),
		Nonce:    HexEncoder.Uint64(tx.Input.Sequence),
		Gas:      HexEncoder.Uint64(tx.GasLimit),
		GasPrice: HexEncoder.Uint64(tx.GasPrice),
		Data:     HexEncoder.Bytes(tx.Data),
	}

	if block != nil {
		// may be pending
		transaction.BlockHash = hexKeccak(block.Hash().Bytes())
		transaction.Hash = HexEncoder.Bytes(hash)
		transaction.BlockNumber = HexEncoder.Uint64(uint64(block.Height))
		transaction.TransactionIndex = hexZero
	}

	if tx.Address != nil {
		transaction.To = HexEncoder.Bytes(tx.Address.Bytes())
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
	d := new(HexDecoder)
	return d.Uint64(height), d.Err()
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
func (srv *EthService) EthSendTransaction(req *EthSendTransactionParams) (*EthSendTransactionResult, error) {
	tx := &payload.CallTx{
		Input: new(payload.TxInput),
	}

	var err error
	d := new(HexDecoder)
	if from := req.Transaction.From; from != "" {
		tx.Input.Address = d.Address(from)
		if d.Err() != nil {
			return nil, fmt.Errorf("failed to parse from address: %v", d.Err())
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

	if to := req.Transaction.To; to != "" {
		addr := d.Address(to)
		if d.Err() != nil {
			return nil, fmt.Errorf("failed to parse to address: %v", d.Err())
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
		bs := d.Bytes(data)
		if d.Err() != nil {
			return nil, fmt.Errorf("failed to parse data: %v", d.Err())
		}
		tx.Data = bs
	}

	txEnv := txs.Enclose(srv.blockchain.ChainID(), tx)

	ctx := context.Background()
	txe, err := srv.trans.BroadcastTxSync(ctx, txEnv)
	if err != nil {
		return nil, err
	} else if txe.Exception != nil {
		return nil, txe.Exception.AsError()
	}

	return &EthSendTransactionResult{
		TransactionHash: HexEncoder.Bytes(txe.GetTxHash().Bytes()),
	}, nil
}

// EthAccounts returns all accounts signable from the local node
func (srv *EthService) EthAccounts() (*EthAccountsResult, error) {
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
		addrs = append(addrs, HexEncoder.Bytes(key.Address.Bytes()))
	}

	return &EthAccountsResult{
		Addresses: addrs,
	}, nil
}

// EthSign: https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_sign
func (srv *EthService) EthSign(req *EthSignParams) (*EthSignResult, error) {
	d := new(HexDecoder)
	to := d.Address(req.Address)
	signer, err := keys.AddressableSigner(srv.keyClient, to)
	if err != nil {
		return nil, err
	}

	data := d.Bytes(req.Bytes)
	if d.Err() != nil {
		return nil, d.Err()
	}

	msg := append([]byte{0x19}, []byte("Ethereum Signed Message:\n")...)
	msg = append(msg, byte(len(data)))
	msg = append(msg, data...)

	sig, err := signer.Sign(crypto.Keccak256(msg))
	if err != nil {
		return nil, err
	}

	return &EthSignResult{
		Signature: HexEncoder.Bytes(sig.RawBytes()),
	}, nil
}

// N / A

func (srv *EthService) EthUninstallFilter(*EthUninstallFilterParams) (*EthUninstallFilterResult, error) {
	return nil, ErrNotFound
}

func (srv *EthService) EthSubmitHashrate(req *EthSubmitHashrateParams) (*EthSubmitHashrateResult, error) {
	return nil, ErrNotFound
}

func (srv *EthService) EthSubmitWork(*EthSubmitWorkParams) (*EthSubmitWorkResult, error) {
	return nil, ErrNotFound
}

func (srv *EthService) EthNewBlockFilter() (*EthNewBlockFilterResult, error) {
	return nil, ErrNotFound
}

func (srv *EthService) EthNewFilter(req *EthNewFilterParams) (*EthNewFilterResult, error) {
	return nil, ErrNotFound
}

func (srv *EthService) EthNewPendingTransactionFilter() (*EthNewPendingTransactionFilterResult, error) {
	return nil, ErrNotFound
}

func (srv *EthService) EthGetUncleByBlockHashAndIndex(req *EthGetUncleByBlockHashAndIndexParams) (*EthGetUncleByBlockHashAndIndexResult, error) {
	return nil, ErrNotFound
}

func (srv *EthService) EthGetUncleByBlockNumberAndIndex(req *EthGetUncleByBlockNumberAndIndexParams) (*EthGetUncleByBlockNumberAndIndexResult, error) {
	return nil, ErrNotFound
}

func (srv *EthService) EthGetUncleCountByBlockHash(req *EthGetUncleCountByBlockHashParams) (*EthGetUncleCountByBlockHashResult, error) {
	return nil, ErrNotFound
}

func (srv *EthService) EthGetUncleCountByBlockNumber(req *EthGetUncleCountByBlockNumberParams) (*EthGetUncleCountByBlockNumberResult, error) {
	return nil, ErrNotFound
}

func (srv *EthService) EthGetProof(req *EthGetProofParams) (*EthGetProofResult, error) {
	return nil, ErrNotFound
}

func (srv *EthService) EthGetWork() (*EthGetWorkResult, error) {
	return nil, ErrNotFound
}

func (srv *EthService) EthGetFilterChanges(req *EthGetFilterChangesParams) (*EthGetFilterChangesResult, error) {
	return nil, ErrNotFound
}

func (srv *EthService) EthGetFilterLogs(req *EthGetFilterLogsParams) (*EthGetFilterLogsResult, error) {
	return nil, ErrNotFound
}

func (srv *EthService) EthCoinbase() (*EthCoinbaseResult, error) {
	return nil, ErrNotFound
}

func (srv *EthService) EthGetLogs(req *EthGetLogsParams) (*EthGetLogsResult, error) {
	return nil, ErrNotFound
}
