package ethclient

import (
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/encoding/web3hex"
	"github.com/hyperledger/burrow/rpc/rpcevents"
)

// These types partially duplicate some of those web3/types.go, the should probably be unified at some point but
// the types in web3/types.go are generated and some are incorrect so adding ones used by client here

type EthLog struct {
	Topics []string `json:"topics"`
	// Hex representation of a Keccak 256 hash
	TransactionHash string `json:"transactionHash"`
	// Sender of the transaction
	Address string `json:"address"`
	// The hex representation of the Keccak 256 of the RLP encoded block
	BlockHash string `json:"blockHash"`
	// The hex representation of the block's height
	BlockNumber string `json:"blockNumber"`
	// Hex representation of a variable length byte array
	Data string `json:"data"`
	// Hex representation of the integer
	LogIndex string `json:"logIndex"`
	// Hex representation of the integer
	TransactionIndex string `json:"transactionIndex"`
}

// This is wrong in web3/types.go
type EthSendTransactionParam struct {
	// Address of the sender
	From string `json:"from"`
	// address of the receiver. null when its a contract creation transaction
	To string `json:"to,omitempty"`
	// The data field sent with the transaction
	Gas string `json:"gas,omitempty"`
	// The gas price willing to be paid by the sender in Wei
	GasPrice string `json:"gasPrice,omitempty"`
	// The gas limit provided by the sender in Wei
	Value string `json:"value,omitempty"`
	// Hex representation of a Keccak 256 hash
	Data string `json:"data"`
	// A number only to be used once
	Nonce string `json:"nonce,omitempty"`
}

type Filter struct {
	*rpcevents.BlockRange
	Addresses []crypto.Address
	Topics    []binary.Word256
}

func (f *Filter) EthFilter() *EthFilter {
	topics := make([]string, len(f.Topics))
	for i, t := range f.Topics {
		topics[i] = web3hex.Encoder.BytesTrim(t[:])
	}
	addresses := make([]string, len(f.Addresses))
	for i, a := range f.Addresses {
		addresses[i] = web3hex.Encoder.Address(a)
	}
	return &EthFilter{
		FromBlock: logBound(f.GetStart()),
		ToBlock:   logBound(f.GetEnd()),
		Addresses: addresses,
		Topics:    topics,
	}
}

type Receipt struct {
	// The hex representation of the block's height
	BlockNumber string `json:"blockNumber"`
	// Hex representation of the integer
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	// Hex representation of the integer
	GasUsed string `json:"gasUsed"`
	// An array of all the logs triggered during the transaction
	Logs []EthLog `json:"logs"`
	// A 2048 bit bloom filter from the logs of the transaction. Each log sets 3 bits though taking the low-order 11 bits of each of the first three pairs of bytes in a Keccak 256 hash of the log's byte series
	TransactionIndex string `json:"transactionIndex"`
	// Whether or not the transaction threw an error.
	Status string `json:"status"`
	// The hex representation of the Keccak 256 of the RLP encoded block
	BlockHash string `json:"blockHash"`
	// The contract address created, if the transaction was a contract creation, otherwise null
	ContractAddress string `json:"contractAddress"`
	// The sender of the transaction
	From string `json:"from"`
	// A 2048 bit bloom filter from the logs of the transaction. Each log sets 3 bits though taking the low-order 11 bits of each of the first three pairs of bytes in a Keccak 256 hash of the log's byte series
	LogsBloom string `json:"logsBloom"`
	// Destination address of the transaction
	To string `json:"to"`
	// Hex representation of a Keccak 256 hash
	TransactionHash string `json:"transactionHash"`
}

// Duplicated here to allow for arrays of addresses
type EthFilter struct {
	// The hex representation of the block's height
	FromBlock string `json:"fromBlock,omitempty"`
	// The hex representation of the block's height
	ToBlock string `json:"toBlock,omitempty"`
	// Yes this is JSON address since allowed to be singular
	Addresses []string `json:"address,omitempty"`
	// Array of 32 Bytes DATA topics. Topics are order-dependent. Each topic can also be an array of DATA with 'or' options
	Topics []string `json:"topics,omitempty"`
}

type Block struct {
	// Hex representation of a Keccak 256 hash
	Sha3Uncles string `json:"sha3Uncles"`
	// Hex representation of a Keccak 256 hash
	TransactionsRoot string `json:"transactionsRoot"`
	// Hex representation of a Keccak 256 hash
	ParentHash string `json:"parentHash"`
	// The address of the beneficiary to whom the mining rewards were given or null when its the pending block
	Miner string `json:"miner"`
	// Integer of the difficulty for this block
	Difficulty string `json:"difficulty"`
	// The total used gas by all transactions in this block
	GasUsed string `json:"gasUsed"`
	// The unix timestamp for when the block was collated
	Timestamp string `json:"timestamp"`
	// Array of transaction objects, or 32 Bytes transaction hashes depending on the last given parameter
	Transactions []string `json:"transactions"`
	// The block number or null when its the pending block
	Number string `json:"number"`
	// The block hash or null when its the pending block
	Hash string `json:"hash"`
	// Array of uncle hashes
	Uncles []string `json:"uncles"`
	// Hex representation of a Keccak 256 hash
	ReceiptsRoot string `json:"receiptsRoot"`
	// The 'extra data' field of this block
	ExtraData string `json:"extraData"`
	// Hex representation of a Keccak 256 hash
	StateRoot string `json:"stateRoot"`
	// Integer of the total difficulty of the chain until this block
	TotalDifficulty string `json:"totalDifficulty"`
	// Integer the size of this block in bytes
	Size string `json:"size"`
	// The maximum gas allowed in this block
	GasLimit string `json:"gasLimit"`
	// Randomly selected number to satisfy the proof-of-work or null when its the pending block
	Nonce string `json:"nonce"`
	// The bloom filter for the logs of the block or null when its the pending block
	LogsBloom string `json:"logsBloom"`
}
