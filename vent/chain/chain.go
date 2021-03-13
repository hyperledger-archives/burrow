// Abstracts over a Burrow GRPC connection and Ethereum json-rpc web3 connection for the purposes of vent

package chain

import (
	"context"
	"time"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/vent/types"
	"google.golang.org/grpc/connectivity"
)

const (
	defaultMaxRetires        = 5
	defaultBackoffBase       = 250 * time.Millisecond
	defaultMaxBlockBatchSize = 100
)

type Chain interface {
	GetChainID() string
	GetVersion() string
	ConsumeBlocks(ctx context.Context, in *rpcevents.BlockRange, consumer func(Block) error) error
	StatusMessage(ctx context.Context, lastProcessedHeight uint64) []interface{}
	Connectivity() connectivity.State
	GetABI(ctx context.Context, address crypto.Address) (string, error)
	Close() error
}

type Block interface {
	GetHeight() uint64
	GetTxs() []Transaction
	GetMetadata(columns types.SQLColumnNames) (map[string]interface{}, error)
}

type Transaction interface {
	GetHash() binary.HexBytes
	GetIndex() uint64
	GetEvents() []Event
	GetException() *errors.Exception
	GetOrigin() *Origin
	GetMetadata(columns types.SQLColumnNames) (map[string]interface{}, error)
}

type Event interface {
	query.Tagged
	GetIndex() uint64
	GetTransactionHash() binary.HexBytes
	GetAddress() crypto.Address
	GetTopics() []binary.Word256
	GetData() []byte
}

type Filter struct {
	Addresses []crypto.Address
	Topics    []binary.Word256
}

type Origin struct {
	// The original ChainID from for this transaction
	ChainID string
	// The original height at which this transaction was committed
	Height uint64
	// The original index in the block
	Index uint64
}

// Client-side block consumer configuration. Requests are retried subject to backoff if a non-fatal error is detected
type BlockConsumerConfig struct {
	// The base backoff - we wait this amount of time between each batch and we increase the backoff exponentially
	// until we reach MaxRetries from BaseBackoffDuration
	BaseBackoffDuration time.Duration
	// The maximum number of retries before failing
	MaxRetries uint64
	// The default and maximum batch size for block requests, we will reduce it logarithmically to a single block
	// when backing off
	MaxBlockBatchSize uint64
}

func (config *BlockConsumerConfig) Complete() {
	if config.MaxBlockBatchSize == 0 {
		config.MaxBlockBatchSize = defaultMaxBlockBatchSize
	}
	if config.BaseBackoffDuration == 0 {
		config.BaseBackoffDuration = defaultBackoffBase
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = defaultMaxRetires
	}
}
