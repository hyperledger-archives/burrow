package service

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/hyperledger/burrow/encoding"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/rpc/lib/jsonrpc"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/web3/ethclient"
	"github.com/hyperledger/burrow/vent/chain"
	"github.com/hyperledger/burrow/vent/chain/burrow"
	"github.com/hyperledger/burrow/vent/chain/ethereum"
	"github.com/hyperledger/burrow/vent/config"
	"github.com/hyperledger/burrow/vent/sqldb"
	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc/connectivity"
)

// Consumer contains basic configuration for consumer to run
type Consumer struct {
	Config *config.VentConfig
	Logger *logging.Logger
	DB     *sqldb.SQLDB
	Chain  chain.Chain
	// external events channel used for when vent is leveraged as a library
	EventsChannel       chan types.EventData
	Done                chan struct{}
	shutdownOnce        sync.Once
	LastProcessedHeight uint64
}

// NewConsumer constructs a new consumer configuration.
// The event channel will be passed a collection of rows generated from all of the events in a single block
// It will be closed by the consumer when it is finished
func NewConsumer(cfg *config.VentConfig, log *logging.Logger, eventChannel chan types.EventData) *Consumer {
	cfg.BlockConsumerConfig.Complete()
	return &Consumer{
		Config:        cfg,
		Logger:        log,
		EventsChannel: eventChannel,
		Done:          make(chan struct{}),
	}
}

// Run connects to a grpc service and subscribes to log events,
// then gets tables structures, maps them & parse event data.
// Store data in SQL event tables, it runs forever
func (c *Consumer) Run(projection *sqlsol.Projection, stream bool) error {
	var err error

	c.Logger.InfoMsg("Connecting to Burrow gRPC server")

	c.Chain, err = c.connectToChain()
	if err != nil {
		return errors.Wrapf(err, "Error connecting to Burrow gRPC server at %s", c.Config.ChainAddress)
	}
	defer c.Chain.Close()
	defer close(c.EventsChannel)

	abiProvider, err := NewAbiProvider(c.Config.AbiFileOrDirs, c.Chain, c.Logger)
	if err != nil {
		return errors.Wrapf(err, "Error loading ABIs")
	}

	if len(projection.Spec) == 0 {
		c.Logger.InfoMsg("No events specifications found")
		return nil
	}

	c.Logger.InfoMsg("Connecting to SQL database")

	connection := types.SQLConnection{
		DBAdapter: c.Config.DBAdapter,
		DBURL:     c.Config.DBURL,
		DBSchema:  c.Config.DBSchema,
		Log:       c.Logger,
	}

	c.DB, err = sqldb.NewSQLDB(connection)
	if err != nil {
		return fmt.Errorf("error connecting to SQL database: %v", err)
	}
	defer c.DB.Close()

	err = c.DB.Init(c.Chain.GetChainID(), c.Chain.GetVersion())
	if err != nil {
		return fmt.Errorf("could not clean tables after ChainID change: %v", err)
	}

	c.Logger.InfoMsg("Synchronizing config and database projection structures")

	err = c.DB.SynchronizeDB(c.Chain.GetChainID(), projection.Tables)
	if err != nil {
		return errors.Wrap(err, "Error trying to synchronize database")
	}

	// doneCh is used for sending a "done" signal from each goroutine to the main thread
	// eventCh is used for sending received events to the main thread to be stored in the db
	errCh := make(chan error, 1)
	eventCh := make(chan types.EventData)

	go func() {
		defer func() {
			c.Shutdown()
		}()
		go c.announceEvery(c.Done)

		c.Logger.InfoMsg("Getting last processed block number from SQL log table")

		fromBlock, err := c.DB.LastBlockHeight(c.Chain.GetChainID())
		if err != nil {
			errCh <- errors.Wrapf(err, "Error trying to get last processed block number")
			return
		}

		startingBlock := fromBlock
		// Start the block after the last one successfully committed - apart from if this is the first block
		// We include block 0 because it is where we currently place dump/restored transactions
		if startingBlock > 0 {
			startingBlock++
		}

		// Allows us skip historical/checkpointed state
		if startingBlock < c.Config.MinimumHeight {
			startingBlock = c.Config.MinimumHeight
		}

		// setup block range to get needed blocks server side
		var end *rpcevents.Bound
		if stream {
			end = rpcevents.StreamBound()
		} else {
			end = rpcevents.LatestBound()
		}

		request := &rpcevents.BlocksRequest{
			BlockRange: rpcevents.NewBlockRange(rpcevents.AbsoluteBound(startingBlock), end),
		}

		c.Logger.TraceMsg("Waiting for blocks...")

		// gets blocks in given range based on last processed block taken from database
		consumer := NewBlockConsumer(c.Chain.GetChainID(), projection, c.Config.SpecOpt, abiProvider.GetEventABI,
			eventCh, c.Done, c.Logger)

		err = c.Chain.ConsumeBlocks(context.Background(), request.BlockRange, consumer)

		if err != nil {
			if err == io.EOF {
				c.Logger.InfoMsg("EOF stream received...")
			} else {
				if finished(c.Done) {
					c.Logger.TraceMsg("GRPC connection closed")
				} else {
					errCh <- errors.Wrapf(err, "Error receiving blocks")
					return
				}
			}
		}
	}()

	for {
		select {
		// Process block events
		case blk := <-eventCh:
			c.LastProcessedHeight = blk.BlockHeight
			err := c.commitBlock(projection, blk)
			if err != nil {
				c.Logger.InfoMsg("error committing block", "err", err)
				return err
			}

		// Await completion
		case <-c.Done:
			select {

			// Select possible error
			case err := <-errCh:
				c.Logger.InfoMsg("finished with error", "err", err)
				return err

			// Or fallback to success
			default:
				c.Logger.InfoMsg("finished successfully")
				return nil
			}
		}
	}
}

func (c *Consumer) commitBlock(projection *sqlsol.Projection, blockEvents types.EventData) error {
	// upsert rows in specific SQL event tables and update block number
	if err := c.DB.SetBlock(c.Chain.GetChainID(), projection.Tables, blockEvents); err != nil {
		return fmt.Errorf("error upserting rows in database: %v", err)
	}

	// send to the external events channel in a non-blocking manner
	select {
	case c.EventsChannel <- blockEvents:
	default:
	}
	return nil
}

// Health returns the health status for the consumer
func (c *Consumer) Health() error {
	if finished(c.Done) {
		return errors.New("closing service")
	}

	// check db status
	if c.DB == nil {
		return errors.New("database disconnected")
	}

	if err := c.DB.Ping(); err != nil {
		return errors.New("database unavailable")
	}

	// check grpc connection status
	if c.Chain == nil {
		return errors.New("grpc disconnected")
	}

	if grpcState := c.Chain.Connectivity(); grpcState != connectivity.Ready {
		return errors.New("grpc connection not ready")
	}

	return nil
}

// Shutdown gracefully shuts down the events consumer
func (c *Consumer) Shutdown() {
	c.shutdownOnce.Do(func() {
		c.Logger.InfoMsg("Shutting down vent consumer...")
		close(c.Done)
		err := c.Chain.Close()
		if err != nil {
			c.Logger.InfoMsg("Could not close Chain connection", structure.ErrorKey, err)
		}
	})
}

func (c *Consumer) StatusMessage(ctx context.Context) []interface{} {
	return c.Chain.StatusMessage(context.Background(), c.LastProcessedHeight)
}

func (c *Consumer) announceEvery(doneCh <-chan struct{}) {
	if c.Config.AnnounceEvery != 0 {
		ticker := time.NewTicker(c.Config.AnnounceEvery)
		for {
			select {
			case <-ticker.C:
				c.Logger.InfoMsg("Announcement", c.StatusMessage(context.Background())...)
			case <-doneCh:
				ticker.Stop()
				return
			}
		}
	}
}

func finished(doneCh chan struct{}) bool {
	select {
	case <-doneCh:
		return true
	default:
		return false
	}
}

func (c *Consumer) connectToChain() (chain.Chain, error) {
	filter := &chain.Filter{
		Addresses: c.Config.WatchAddresses,
	}
	c.Logger.InfoMsg("Attempting to detect chain type", "chain_address", c.Config.ChainAddress)
	burrowChain, burrowErr := dialBurrow(c.Config.ChainAddress, filter)
	if burrowErr == nil {
		return burrowChain, nil
	}
	ethChain, ethErr := dialEthereum(c.Config.ChainAddress, filter, &c.Config.BlockConsumerConfig, c.Logger)
	if ethErr != nil {
		return nil, fmt.Errorf("could not connect to either Burrow or Ethereum chain, "+
			"Burrow error: %v, Ethereum error: %v", burrowErr, ethErr)
	}
	return ethChain, nil
}

func dialBurrow(chainAddress string, filter *chain.Filter) (*burrow.Chain, error) {
	conn, err := encoding.GRPCDial(chainAddress)
	if err != nil {
		return nil, err
	}
	return burrow.New(conn, filter)
}

func dialEthereum(chainAddress string, filter *chain.Filter, consumerConfig *chain.BlockConsumerConfig,
	logger *logging.Logger) (*ethereum.Chain, error) {
	client := ethclient.NewEthClient(jsonrpc.NewClient(chainAddress))
	return ethereum.New(client, filter, consumerConfig, logger)
}
