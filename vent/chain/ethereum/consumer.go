package ethereum

import (
	"bytes"
	"fmt"
	"time"

	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/rpc/lib/types"
	"github.com/pkg/errors"

	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/web3/ethclient"
	"github.com/hyperledger/burrow/vent/chain"
)

const (
	defaultMaxRetires = 5
	defaultMaxBlocks  = 1000
	backoffBase       = 10 * time.Millisecond
)

type consumer struct {
	client     rpc.Client
	filter     *chain.Filter
	blockRange *rpcevents.BlockRange
	logger     *logging.Logger
	consumer   func(block chain.Block) error
	// Next unconsumed height
	nextBlockHeight   uint64
	retries           uint64
	backoffDuration   time.Duration
	maxRetries        uint64
	maxBlockBatchSize uint64
	blockBatchSize    uint64
}

func Consume(client rpc.Client, filter *chain.Filter, blockRange *rpcevents.BlockRange, logger *logging.Logger,
	consume func(block chain.Block) error) error {
	c := consumer{
		client:            client,
		filter:            filter,
		blockRange:        blockRange,
		logger:            logger,
		consumer:          consume,
		backoffDuration:   backoffBase,
		maxRetries:        defaultMaxRetires,
		maxBlockBatchSize: defaultMaxBlocks,
		blockBatchSize:    defaultMaxBlocks,
	}
	return c.Consume()
}

func (c *consumer) Consume() error {
	start, end, streaming, err := c.bounds()
	if err != nil {
		return err
	}

	for c.nextBlockHeight <= end || streaming {
		err = c.ConsumeInBatches(start, end)
		if err != nil {
			return err
		}
		start, end, streaming, err = c.bounds()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *consumer) ConsumeInBatches(start, end uint64) error {
	for batchStart := start; batchStart < end; batchStart += c.blockBatchSize {
		batchEnd := batchStart + c.blockBatchSize
		if batchEnd > end {
			batchEnd = end
		}
		logs, err := ethclient.EthGetLogs(c.client, &ethclient.Filter{
			BlockRange: rpcevents.AbsoluteRange(batchStart, batchEnd),
			Addresses:  c.filter.Addresses,
			Topics:     c.filter.Topics,
		})
		if err != nil {
			err = c.handleError(end, err)
			if err != nil {
				return err
			}
			// We managed to handle the error (a retry was successful)
			return nil
		}
		// Request was successful
		c.recover()
		lastBlock, err := consumeBlocksFromLogs(logs, c.consumer)
		if err != nil {
			return fmt.Errorf("could not consume ethereum logs: %w", err)
		}
		if lastBlock != nil {
			c.nextBlockHeight = lastBlock.GetHeight() + 1
		}
	}
	return nil
}

func (c *consumer) bounds() (start uint64, end uint64, streaming bool, err error) {
	var latestHeight uint64

	latestHeight, err = ethclient.EthBlockNumber(c.client)
	if err != nil {
		err = fmt.Errorf("could not get latest height: %w", err)
		return
	}
	start, end, streaming = c.blockRange.Bounds(latestHeight)

	if start < c.nextBlockHeight {
		start = c.nextBlockHeight
	}
	return
}

func (c *consumer) handleError(end uint64, err error) error {
	var rpcError *types.RPCError
	if errors.As(err, &rpcError) {
		// If we have a custom server error maybe our batch size is too large or maybe we should wait
		if rpcError.IsServerError() {
			c.retries++
			if c.retries <= c.maxRetries {
				// Server may throw if batch too large or request takes too long
				c.backoff()
				c.logger.InfoMsg("Ethereum block consumer retrying after Ethereum Server Error", structure.ErrorKey, rpcError)
				return c.ConsumeInBatches(c.nextBlockHeight, end)
			}
		}
	}
	return err
}

// Asymptotic decrease to single block
func (c *consumer) backoff() {
	c.blockBatchSize /= 2
	if c.blockBatchSize == 0 {
		c.blockBatchSize = 1
	}
	time.Sleep(c.backoffDuration)
	c.backoffDuration *= 2
}

// Asymptotic increase to max blocks
func (c *consumer) recover() {
	delta := (c.maxBlockBatchSize - c.blockBatchSize) / 2
	if delta == 0 {
		c.blockBatchSize = c.maxBlockBatchSize
	} else {
		c.blockBatchSize += delta
	}
	// Reset retries and backoff
	c.backoffDuration = backoffBase
	c.retries = 0
}

func consumeBlocksFromLogs(logs []*ethclient.EthLog, consumer func(block chain.Block) error) (chain.Block, error) {
	if len(logs) == 0 {
		return nil, nil
	}
	log, err := NewEthereumEvent(logs[0])
	if err != nil {
		return nil, fmt.Errorf("could not deserialise ethereum event: %w", err)
	}
	block := NewEthereumBlock(log)
	txHash := log.TransactionHash
	indexInBlock := log.IndexInBlock

	for i := 1; i < len(logs); i++ {
		log, err = NewEthereumEvent(logs[i])
		if err != nil {
			return nil, fmt.Errorf("could not deserialise ethereum event: %w", err)
		}
		if log.Height > block.Height {
			// New block
			err = consumer(block)
			if err != nil {
				return nil, err
			}
			// Establish new block
			block = NewEthereumBlock(log)
		} else {
			if log.IndexInBlock <= indexInBlock {
				return nil, fmt.Errorf("event LogIndex is non-increasing within block, "+
					"previous LogIndex was %d but current is %d (at height %d)", indexInBlock, log.IndexInBlock, block.Height)
			}
			if !bytes.Equal(txHash, log.TransactionHash) {
				// New Tx
				block.appendTransaction(log)
			} else {
				block.appendEvent(log)
			}
		}
		txHash = log.TransactionHash
		indexInBlock = log.IndexInBlock
	}
	return block, consumer(block)
}
