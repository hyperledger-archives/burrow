package rpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/acm/validator"
	bcm "github.com/hyperledger/burrow/bcm"
	bin "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/encoding"
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
	data, err := encoding.HexDecodeToBytes(req.Data)
	if err != nil {
		return nil, err
	}

	hash, err := crypto.LegacyKeccak256Hash(data)
	if err != nil {
		return nil, err
	}

	return &web3.Web3Sha3Result{
		HashedData: encoding.HexEncodeBytes(hash),
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
		NumConnectedPeers: encoding.HexEncodeNumber(uint64(srv.nodeView.Peers().Size())),
	}, nil
}

// NetVersion returns the hex encoding of the genesis hash,
// this is typically a small int (where 1 == Ethereum mainnet)
func (srv *EthService) NetVersion() (*web3.NetVersionResult, error) {
	doc := srv.blockchain.GenesisDoc()
	return &web3.NetVersionResult{
		ChainID: encoding.HexEncodeBytes(doc.ShortHash()),
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
		BlockNumber: encoding.HexEncodeNumber(srv.blockchain.LastBlockHeight()),
	}, nil
}

// EthCall executes a new message call immediately without creating a transaction
func (srv *EthService) EthCall(req *web3.EthCallParams) (*web3.EthCallResult, error) {
	var to, from crypto.Address
	var err error

	if addr := req.Transaction.To; addr != "" {
		to, err = encoding.HexDecodeToAddress(addr)
		if err != nil {
			return nil, err
		}
	}

	if addr := req.Transaction.From; addr != "" {
		from, err = encoding.HexDecodeToAddress(addr)
		if err != nil {
			return nil, err
		}
	}

	// don't hex decode abi
	data := []byte(encoding.HexRemovePrefix(req.Transaction.Data))

	txe, err := execution.CallSim(srv.accounts, srv.blockchain, from, to, data, srv.logger)
	if err != nil {
		return nil, err
	}

	var result string
	if r := txe.GetResult(); r != nil {
		result = encoding.HexEncodeBytes(r.GetReturn())
	}

	return &web3.EthCallResult{
		ReturnValue: result,
	}, nil
}

// EthGetBalance returns an accounts balance, or an error if it does not exist
func (srv *EthService) EthGetBalance(req *web3.EthGetBalanceParams) (*web3.EthGetBalanceResult, error) {
	addr, err := encoding.HexDecodeToAddress(req.Address)
	if err != nil {
		return nil, err
	}

	acc, err := srv.accounts.GetAccount(addr)
	if err != nil {
		return nil, err
	} else if acc == nil {
		return nil, fmt.Errorf("account not found at address %s", req.Address)
	}

	return &web3.EthGetBalanceResult{
		GetBalanceResult: encoding.HexEncodeNumber(acc.Balance),
	}, nil
}

// EthGetBlockByHash iterates through all headers to find a matching block height for a given hash
func (srv *EthService) EthGetBlockByHash(req *web3.EthGetBlockByHashParams) (*web3.EthGetBlockByHashResult, error) {
	height, err := srv.getBlockHeightByHash(req.BlockHash)
	if err != nil {
		return nil, err
	}

	block, err := srv.getBlockInfoAtHeight(height)
	if err != nil {
		return nil, err
	}

	return &web3.EthGetBlockByHashResult{
		GetBlockByHashResult: web3.GetBlockByHashResult{
			Block: *block,
		},
	}, web3.ErrNotFound
}

// EthGetBlockByNumber returns block info at the given height
func (srv *EthService) EthGetBlockByNumber(req *web3.EthGetBlockByNumberParams) (*web3.EthGetBlockByNumberResult, error) {
	height, err := srv.getHeightByWordOrNumber(req.BlockNumber)
	if err != nil {
		return nil, err
	}

	block, err := srv.getBlockInfoAtHeight(height)
	if err != nil {
		return nil, err
	}

	return &web3.EthGetBlockByNumberResult{
		GetBlockByNumberResult: web3.GetBlockByNumberResult{
			Block: *block,
		},
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
		BlockTransactionCountByHash: encoding.HexEncodeNumber(uint64(block.NumTxs)),
	}, web3.ErrNotFound
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
		BlockTransactionCountByHash: encoding.HexEncodeNumber(uint64(block.NumTxs)),
	}, nil
}

// EthGetCode returns the EVM bytecode at an address
func (srv *EthService) EthGetCode(req *web3.EthGetCodeParams) (*web3.EthGetCodeResult, error) {
	addr, err := encoding.HexDecodeToAddress(req.Address)
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
		Bytes: encoding.HexEncodeBytes(acc.EVMCode),
	}, nil
}

// TODO
func (srv *EthService) EthGetRawTransactionByHash(req *web3.EthGetRawTransactionByHashParams) (*web3.EthGetRawTransactionByHashResult, error) {
	return nil, web3.ErrNotFound
}

// TODO
func (srv *EthService) EthGetRawTransactionByBlockHashAndIndex(req *web3.EthGetRawTransactionByBlockHashAndIndexParams) (*web3.EthGetRawTransactionByBlockHashAndIndexResult, error) {
	return nil, web3.ErrNotFound
}

// TODO
func (srv *EthService) EthGetRawTransactionByBlockNumberAndIndex(req *web3.EthGetRawTransactionByBlockNumberAndIndexParams) (*web3.EthGetRawTransactionByBlockNumberAndIndexResult, error) {
	return nil, web3.ErrNotFound
}

// TODO
func (srv *EthService) EthGetStorageAt(req *web3.EthGetStorageAtParams) (*web3.EthGetStorageAtResult, error) {
	// addr, err := crypto.AddressFromHexString(req.Address)
	// if err != nil {
	// 	return nil, err
	// }

	return nil, web3.ErrNotFound
}

// TODO
func (srv *EthService) EthGetTransactionByBlockHashAndIndex(req *web3.EthGetTransactionByBlockHashAndIndexParams) (*web3.EthGetTransactionByBlockHashAndIndexResult, error) {
	return nil, web3.ErrNotFound
}

// TODO
func (srv *EthService) EthGetTransactionByBlockNumberAndIndex(req *web3.EthGetTransactionByBlockNumberAndIndexParams) (*web3.EthGetTransactionByBlockNumberAndIndexResult, error) {
	return nil, web3.ErrNotFound
}

// EthGetTransactionByHash finds a tx by the given hash
func (srv *EthService) EthGetTransactionByHash(req *web3.EthGetTransactionByHashParams) (*web3.EthGetTransactionByHashResult, error) {
	hash, err := encoding.HexDecodeToBytes(req.TransactionHash)
	if err != nil {
		return nil, err
	}

	txe, err := srv.events.TxByHash(hash)
	if err != nil {
		return nil, err
	} else if txe.Envelope == nil {
		return nil, fmt.Errorf("no envelope for tx %s", req.TransactionHash)
	} else if txe.Envelope.Tx == nil {
		return nil, fmt.Errorf("no payload for tx %s", req.TransactionHash)
	}

	head, err := srv.blockchain.GetBlockHeader(txe.Height)
	if err != nil {
		return nil, err
	}

	tx := payloadToTx(txe.Envelope.Tx.Payload)
	tx.Hash = encoding.HexEncodeBytes(txe.GetTxHash())
	tx.BlockNumber = encoding.HexEncodeBytes([]byte(strconv.FormatUint(txe.Height, 10)))
	tx.BlockHash = encoding.HexEncodeBytes(head.Hash())

	return &web3.EthGetTransactionByHashResult{
		Transaction: tx,
	}, nil
}

// EthGetTransactionCount returns the number of transactions sent from an address
func (srv *EthService) EthGetTransactionCount(req *web3.EthGetTransactionCountParams) (*web3.EthGetTransactionCountResult, error) {
	// TODO: implement
	return nil, web3.ErrNotFound
}

// EthGetTransactionReceipt returns the receipt of a previously committed tx
func (srv *EthService) EthGetTransactionReceipt(req *web3.EthGetTransactionReceiptParams) (*web3.EthGetTransactionReceiptResult, error) {
	data, err := encoding.HexDecodeToBytes(req.TransactionHash)
	if err != nil {
		return nil, err
	}

	txe, err := srv.events.TxByHash(data)
	if err != nil {
		return nil, err
	} else if txe == nil {
		return nil, fmt.Errorf("tx with hash %s does not exist", req.TransactionHash)
	}

	block, err := srv.blockchain.GetBlockHeader(txe.Height)
	if err != nil {
		return nil, err
	}

	result := &web3.EthGetTransactionReceiptResult{
		Receipt: web3.Receipt{
			Status:      "0x1",
			BlockNumber: encoding.HexEncodeNumber(uint64(block.Height)),
			BlockHash:   encoding.HexEncodeBytes(block.Hash()),
			From:        combineInputs(txe.Envelope.Tx.GetInputs()...),
			// CumulativeGasUsed
			GasUsed:         encoding.HexEncodeNumber(txe.Result.GetGasUsed()),
			TransactionHash: encoding.HexEncodeBytes(txe.GetTxHash()),
			// TransactionIndex: strconv.FormatUint(txe.GetIndex(), 10),
		},
	}

	if txe.Receipt != nil {
		// hex encoded
		result.Receipt.ContractAddress = encoding.HexEncodeBytes(txe.Receipt.ContractAddress.Bytes())
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
		pending = append(pending, web3.PendingTransactions{
			Transaction: payloadToTx(env.Tx.Payload),
		})
	}

	return &web3.EthPendingTransactionsResult{
		PendingTransactions: pending,
	}, nil
}

type RawTx struct {
	Nonce    uint64 `json:"nonce"`
	GasPrice uint64 `json:"gasPrice"`
	Gas      uint64 `json:"gas"`
	To       []byte `json:"to"`
	Value    uint64 `json:"value"`
	Data     []byte `json:"data"`

	V uint64 `json:"v"`
	R []byte `json:"r"`
	S []byte `json:"s"`
}

func (srv *EthService) EthSendRawTransaction(req *web3.EthSendRawTransactionParams) (*web3.EthSendRawTransactionResult, error) {
	data, err := encoding.HexDecodeToBytes(req.SignedTransactionData)
	if err != nil {
		return nil, err
	}

	rawTx := new(RawTx)
	err = rlp.Decode(data, rawTx)
	if err != nil {
		return nil, err
	}

	toHash := []interface{}{
		rawTx.Nonce,
		rawTx.GasPrice,
		rawTx.Gas,
		rawTx.To,
		rawTx.Value,
		rawTx.Data,
		big.NewInt(1).Uint64(), uint(0), uint(0),
	}

	enc, err := rlp.Encode(toHash)
	if err != nil {
		return nil, err
	}

	hash, err := crypto.LegacyKeccak256Hash(enc)
	if err != nil {
		return nil, err
	}

	sig := crypto.CompressedSignatureFromParams(rawTx.V-1, rawTx.R, rawTx.S)
	pub, err := crypto.PublicKeyFromSignature(sig, hash)
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
			Payload: &payload.SendTx{
				Inputs: []*payload.TxInput{
					{
						Address:  from,
						Amount:   rawTx.Value,
						Sequence: rawTx.Nonce,
					},
				},
				Outputs: []*payload.TxOutput{
					{
						Address: to,
						Amount:  rawTx.Value,
					},
				},
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

	return nil, nil
}

// EthSyncing returns this nodes syncing status (i.e. whether it has caught up)
func (srv *EthService) EthSyncing() (*web3.EthSyncingResult, error) {
	// TODO: remaining sync fields
	return &web3.EthSyncingResult{
		SyncStatus: web3.SyncStatus{
			CurrentBlock: encoding.HexEncodeNumber(srv.blockchain.LastBlockHeight()),
		},
	}, nil
}

func (srv *EthService) getBlockHeightByHash(hash string) (uint64, error) {
	for i := uint64(0); i < srv.blockchain.LastBlockHeight(); i++ {
		head, err := srv.blockchain.GetBlockHeader(i)
		if err != nil {
			return 0, err
		} else if head.Hash().String() == hash {
			return i, nil
		}
	}
	return 0, web3.ErrNotFound
}

func (srv *EthService) getBlockHeaderAtHeight(height uint64) (*types.Header, error) {
	return srv.blockchain.GetBlockHeader(height)
}

func (srv *EthService) getBlockInfoAtHeight(height uint64) (*web3.Block, error) {
	block, err := srv.getBlockHeaderAtHeight(height)
	if err != nil {
		return nil, err
	} else if block == nil {
		return nil, fmt.Errorf("block at height %d does not exist", height)
	}

	return &web3.Block{
		TransactionsRoot: encoding.HexEncodeBytes(block.AppHash.Bytes()),
	}, nil
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
	return encoding.HexDecodeToNumber(height)
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

func combineInputs(ins ...*payload.TxInput) string {
	addrs := make([]string, 0, len(ins))
	for _, i := range ins {
		addrs = append(addrs, encoding.HexEncodeBytes(i.Address.Bytes()))
	}
	return strings.Join(addrs, ",")
}

func combineOutputs(outs ...*payload.TxOutput) string {
	addrs := make([]string, 0, len(outs))
	for _, o := range outs {
		addrs = append(addrs, encoding.HexEncodeBytes(o.Address.Bytes()))
	}
	return strings.Join(addrs, ",")
}

func payloadToTx(in payload.Payload) web3.Transaction {
	switch tx := in.(type) {
	case *payload.CallTx:
		transaction := web3.Transaction{
			From:  encoding.HexEncodeBytes(tx.Input.Address.Bytes()),
			Value: encoding.HexEncodeNumber(tx.Input.Amount),
			Nonce: encoding.HexEncodeNumber(tx.Input.Sequence),
			Gas:   encoding.HexEncodeNumber(tx.GasLimit),
			Data:  encoding.HexEncodeBytes(tx.Data),
		}

		if tx.Address != nil {
			transaction.To = encoding.HexEncodeBytes(tx.Address.Bytes())
		} else {
			transaction.To = "null"
		}

		return transaction
	case *payload.SendTx:
		return web3.Transaction{
			From: combineInputs(tx.Inputs...),
			To:   combineOutputs(tx.Outputs...),
		}
	default:
		return web3.Transaction{
			From: combineInputs(tx.GetInputs()...),
		}
	}
}

// TODO: deprecate? > https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1767.md#rationale

type SendOrCall struct {
	input payload.TxInput
	to    *crypto.Address
	gas   uint64
}

// EthSendTransaction constructs, signs and broadcasts a tx from the local node
func (srv *EthService) EthSendTransaction(req *web3.EthSendTransactionParams) (*web3.EthSendTransactionResult, error) {
	tx := SendOrCall{}

	var err error
	if from := req.Transaction.From; from != "" {
		tx.input.Address, err = encoding.HexDecodeToAddress(from)
		if err != nil {
			return nil, fmt.Errorf("failed to parse from address: %v", err)
		}
	} else {
		return nil, fmt.Errorf("no from address specified")
	}

	if value := req.Transaction.Value; value != "" {
		tx.input.Amount, err = strconv.ParseUint(value, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount: %v", err)
		}
	}

	acc, err := srv.accounts.GetAccount(tx.input.Address)
	if err != nil {
		return nil, err
	}

	tx.input.Sequence = acc.Sequence + 1

	if to := req.Transaction.To; to != "" {
		addr, err := encoding.HexDecodeToAddress(to)
		if err != nil {
			return nil, fmt.Errorf("failed to parse to address: %v", err)
		}
		tx.to = &addr
	}

	// gas provided for the transaction execution
	if gas := req.Transaction.Gas; gas != "" {
		tx.gas, err = strconv.ParseUint(gas, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse gas: %v", err)
		}
	}

	var input payload.Payload
	if data := req.Transaction.Data; data != "" {
		bs, err := encoding.HexDecodeToBytes(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse data: %v", err)
		}
		// if request contains data then do a callTx
		input = &payload.CallTx{
			Input:    &tx.input,
			Address:  tx.to,
			GasLimit: tx.gas,
			Data:     bin.HexBytes(bs),
		}
	} else {
		input = &payload.SendTx{
			Inputs: []*payload.TxInput{&tx.input},
			Outputs: []*payload.TxOutput{
				&payload.TxOutput{
					Amount: tx.input.Amount,
				},
			},
		}
	}

	txEnv := txs.Enclose(srv.blockchain.ChainID(), input)

	signer, err := keys.AddressableSigner(srv.keyClient, tx.input.Address)
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
		TransactionHash: encoding.HexEncodeBytes(txe.GetTxHash().Bytes()),
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
		}
		addrs = append(addrs, encoding.HexEncodeBytes(key.Address.Bytes()))
	}

	return &web3.EthAccountsResult{
		Addresses: addrs,
	}, nil
}

func (srv *EthService) EthSign(req *web3.EthSignParams) (*web3.EthSignResult, error) {
	return nil, web3.ErrNotFound
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

func (srv *EthService) EthEstimateGas(req *web3.EthEstimateGasParams) (*web3.EthEstimateGasResult, error) {
	return nil, web3.ErrNotFound
}

func (srv *EthService) EthGasPrice() (*web3.EthGasPriceResult, error) {
	return nil, web3.ErrNotFound
}
