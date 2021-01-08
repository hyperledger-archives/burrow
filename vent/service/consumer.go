package service

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/hyperledger/burrow/encoding"

	"github.com/hyperledger/burrow/rpc"

	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/hyperledger/burrow/vent/config"
	"github.com/hyperledger/burrow/vent/sqldb"
	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// Consumer contains basic configuration for consumer to run
type Consumer struct {
	Config         *config.VentConfig
	Logger         *logging.Logger
	DB             *sqldb.SQLDB
	GRPCConnection *grpc.ClientConn
	// external events channel used for when vent is leveraged as a library
	EventsChannel chan types.EventData
	Done          chan struct{}
	shutdownOnce  sync.Once
	Status
}

// Status announcement
type Status struct {
	LastProcessedHeight uint64
	Burrow              *rpc.ResultStatus
}

// NewConsumer constructs a new consumer configuration.
// The event channel will be passed a collection of rows generated from all of the events in a single block
// It will be closed by the consumer when it is finished
func NewConsumer(cfg *config.VentConfig, log *logging.Logger, eventChannel chan types.EventData) *Consumer {
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

	c.GRPCConnection, err = encoding.GRPCDial(c.Config.GRPCAddr)
	if err != nil {
		return errors.Wrapf(err, "Error connecting to Burrow gRPC server at %s", c.Config.GRPCAddr)
	}
	defer c.GRPCConnection.Close()
	defer close(c.EventsChannel)

	// get the chain ID to compare with the one stored in the db
	qCli := rpcquery.NewQueryClient(c.GRPCConnection)
	c.Status.Burrow, err = qCli.Status(context.Background(), &rpcquery.StatusParam{})
	if err != nil {
		return errors.Wrapf(err, "Error getting chain status")
	}

	abiProvider, err := NewAbiProvider(c.Config.AbiFileOrDirs, rpcquery.NewQueryClient(c.GRPCConnection), c.Logger)
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

	err = c.DB.Init(c.Burrow.ChainID, c.Burrow.BurrowVersion)
	if err != nil {
		return fmt.Errorf("could not clean tables after ChainID change: %v", err)
	}

	c.Logger.InfoMsg("Synchronizing config and database projection structures")

	err = c.DB.SynchronizeDB(c.Burrow.ChainID, projection.Tables)
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

		// NOTE [Silas]: I am preserving the comment below that dates from the early days of Vent. I have looked at the
		// bosmarmot git history and I cannot see why the original author thought that it was the case that there was
		// no way of knowing if the last block of events was committed since the block and its associated log is
		// committed atomically in a transaction and this is a core part of he design of Vent - in order that it does not
		// repeat

		// [ORIGINAL COMMENT]
		// right now there is no way to know if the last block of events was completely read
		// so we have to begin processing from the last block number stored in database
		// and update event data if already present
		fromBlock, err := c.DB.LastBlockHeight(c.Burrow.ChainID)
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

		// setup block range to get needed blocks server side
		cli := rpcevents.NewExecutionEventsClient(c.GRPCConnection)
		var end *rpcevents.Bound
		if stream {
			end = rpcevents.StreamBound()
		} else {
			end = rpcevents.LatestBound()
		}

		request := &rpcevents.BlocksRequest{
			BlockRange: rpcevents.NewBlockRange(rpcevents.AbsoluteBound(startingBlock), end),
		}

		// gets blocks in given range based on last processed block taken from database
		stream, err := cli.Stream(context.Background(), request)
		if err != nil {
			errCh <- errors.Wrapf(err, "Error connecting to block stream")
			return
		}

		// get blocks

		c.Logger.TraceMsg("Waiting for blocks...")

		err = rpcevents.ConsumeBlockExecutions(stream,
			NewBlockConsumer(projection, c.Config.SpecOpt, abiProvider.GetEventAbi, eventCh, c.Done, c.Logger))

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
			c.Status.LastProcessedHeight = blk.BlockHeight
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
	if err := c.DB.SetBlock(c.Burrow.ChainID, projection.Tables, blockEvents); err != nil {
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
	if c.GRPCConnection == nil {
		return errors.New("grpc disconnected")
	}

	if grpcState := c.GRPCConnection.GetState(); grpcState != connectivity.Ready {
		return errors.New("grpc connection not ready")
	}

	return nil
}

// Shutdown gracefully shuts down the events consumer
func (c *Consumer) Shutdown() {
	c.shutdownOnce.Do(func() {
		c.Logger.InfoMsg("Shutting down vent consumer...")
		close(c.Done)
		c.GRPCConnection.Close()
	})
}

func (c *Consumer) updateStatus(qcli rpcquery.QueryClient) {
	stat, err := qcli.Status(context.Background(), &rpcquery.StatusParam{})
	if err != nil {
		c.Logger.InfoMsg("could not get blockchain status", "err", err)
		return
	}
	c.Status.Burrow = stat
}

func (c *Consumer) statusMessage() []interface{} {
	var catchUpRatio float64
	if c.Burrow.SyncInfo.LatestBlockHeight > 0 {
		catchUpRatio = float64(c.LastProcessedHeight) / float64(c.Burrow.SyncInfo.LatestBlockHeight)
	}
	return []interface{}{
		"msg", "status",
		"last_processed_height", c.LastProcessedHeight,
		"fraction_caught_up", catchUpRatio,
		"burrow_latest_block_height", c.Burrow.SyncInfo.LatestBlockHeight,
		"burrow_latest_block_duration", c.Burrow.SyncInfo.LatestBlockDuration,
		"burrow_latest_block_hash", c.Burrow.SyncInfo.LatestBlockHash,
		"burrow_latest_app_hash", c.Burrow.SyncInfo.LatestAppHash,
		"burrow_latest_block_time", c.Burrow.SyncInfo.LatestBlockTime,
		"burrow_latest_block_seen_time", c.Burrow.SyncInfo.LatestBlockSeenTime,
		"burrow_node_info", c.Burrow.NodeInfo,
		"burrow_catching_up", c.Burrow.CatchingUp,
	}
}

func (c *Consumer) announceEvery(doneCh <-chan struct{}) {
	if c.Config.AnnounceEvery != 0 {
		qcli := rpcquery.NewQueryClient(c.GRPCConnection)
		ticker := time.NewTicker(c.Config.AnnounceEvery)
		for {
			select {
			case <-ticker.C:
				c.updateStatus(qcli)
				c.Logger.InfoMsg("Announcement", c.statusMessage()...)
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
