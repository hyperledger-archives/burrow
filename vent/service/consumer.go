package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/hyperledger/burrow/rpc"

	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/exec"
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
	Log            *logging.Logger
	Closing        bool
	DB             *sqldb.SQLDB
	GRPCConnection *grpc.ClientConn
	// external events channel used for when vent is leveraged as a library
	EventsChannel chan types.EventData
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
		Log:           log,
		Closing:       false,
		EventsChannel: eventChannel,
	}
}

// Run connects to a grpc service and subscribes to log events,
// then gets tables structures, maps them & parse event data.
// Store data in SQL event tables, it runs forever
func (c *Consumer) Run(projection *sqlsol.Projection, abiSpec *abi.Spec, stream bool) error {
	var err error

	c.Log.InfoMsg("Connecting to Burrow gRPC server")

	c.GRPCConnection, err = grpc.Dial(c.Config.GRPCAddr, grpc.WithInsecure())
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

	if len(projection.EventSpec) == 0 {
		c.Log.InfoMsg("No events specifications found")
		return nil
	}

	c.Log.InfoMsg("Connecting to SQL database")

	connection := types.SQLConnection{
		DBAdapter: c.Config.DBAdapter,
		DBURL:     c.Config.DBURL,
		DBSchema:  c.Config.DBSchema,
		Log:       c.Log,
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

	c.Log.InfoMsg("Synchronizing config and database projection structures")

	err = c.DB.SynchronizeDB(c.Burrow.ChainID, projection.Tables)
	if err != nil {
		return errors.Wrap(err, "Error trying to synchronize database")
	}

	// doneCh is used for sending a "done" signal from each goroutine to the main thread
	// eventCh is used for sending received events to the main thread to be stored in the db
	doneCh := make(chan struct{})
	errCh := make(chan error, 1)
	eventCh := make(chan types.EventData)

	go func() {
		defer func() {
			close(doneCh)
		}()
		go c.announceEvery(doneCh)

		c.Log.InfoMsg("Getting last processed block number from SQL log table")

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

		c.Log.TraceMsg("Waiting for blocks...")

		err = rpcevents.ConsumeBlockExecutions(stream, c.makeBlockConsumer(projection, abiSpec, eventCh))

		if err != nil {
			if err == io.EOF {
				c.Log.InfoMsg("EOF stream received...")
			} else {
				if c.Closing {
					c.Log.TraceMsg("GRPC connection closed")
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
			err := c.commitBlock(projection, blk)
			if err != nil {
				c.Log.InfoMsg("error committing block", "err", err)
				return err
			}

		// Await completion
		case <-doneCh:
			select {

			// Select possible error
			case err := <-errCh:
				c.Log.InfoMsg("finished with error", "err", err)
				return err

			// Or fallback to success
			default:
				c.Log.InfoMsg("finished successfully")
				return nil
			}
		}
	}
}

func (c *Consumer) makeBlockConsumer(projection *sqlsol.Projection, abiSpec *abi.Spec,
	eventCh chan<- types.EventData) func(blockExecution *exec.BlockExecution) error {

	return func(blockExecution *exec.BlockExecution) error {
		if c.Closing {
			return io.EOF
		}

		// set new block number
		fromBlock := blockExecution.Height

		defer func() {
			c.Status.LastProcessedHeight = fromBlock
		}()

		c.Log.TraceMsg("Block received", "height", blockExecution.Height, "num_txs", len(blockExecution.TxExecutions))

		// create a fresh new structure to store block data at this height
		blockData := sqlsol.NewBlockData(fromBlock)

		if c.Config.SpecOpt&sqlsol.Block > 0 {
			blkRawData, err := buildBlkData(projection.Tables, blockExecution)
			if err != nil {
				return errors.Wrapf(err, "Error building block raw data")
			}
			// set row in structure
			blockData.AddRow(tables.Block, blkRawData)
		}

		// get transactions for a given block
		for _, txe := range blockExecution.TxExecutions {
			c.Log.TraceMsg("Getting transaction", "TxHash", txe.TxHash, "num_events", len(txe.Events))

			if c.Config.SpecOpt&sqlsol.Tx > 0 {
				txRawData, err := buildTxData(txe)
				if err != nil {
					return errors.Wrapf(err, "Error building tx raw data")
				}
				// set row in structure
				blockData.AddRow(tables.Tx, txRawData)
			}

			// reverted transactions don't have to update event data tables
			// so check that condition to filter them
			if txe.Exception == nil {

				origin := txe.Origin
				if origin == nil {
					origin = &exec.Origin{
						ChainID: c.Burrow.ChainID,
						Height:  txe.Height,
					}
				}

				// get events for a given transaction
				for _, event := range txe.Events {

					taggedEvent := event.Tagged()

					// see which spec filter matches with the one in event data
					for _, eventClass := range projection.EventSpec {
						qry, err := eventClass.Query()

						if err != nil {
							return errors.Wrapf(err, "Error parsing query from filter string")
						}

						// there's a matching filter, add data to the rows
						if qry.Matches(taggedEvent) {

							c.Log.InfoMsg("Matched event", "header", event.Header,
								"filter", eventClass.Filter)

							// unpack, decode & build event data
							eventData, err := buildEventData(projection, eventClass, event, origin, abiSpec, c.Log)
							if err != nil {
								return errors.Wrapf(err, "Error building event data")
							}

							// set row in structure
							blockData.AddRow(eventClass.TableName, eventData)
						}
					}
				}
			}
		}

		// upsert rows in specific SQL event tables and update block number
		// store block data in SQL tables (if any)
		if blockData.PendingRows(fromBlock) {
			// gets block data to upsert
			blk := blockData.Data

			for name, rows := range blk.Tables {
				c.Log.InfoMsg("Upserting rows in SQL table", "height", fromBlock, "table", name, "action", "UPSERT", "rows", rows)
			}

			eventCh <- blk
		}
		return nil
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
	if c.Closing {
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
	c.Log.InfoMsg("Shutting down vent consumer...")
	c.Closing = true
	c.GRPCConnection.Close()
}

func (c *Consumer) updateStatus(qcli rpcquery.QueryClient) {
	stat, err := qcli.Status(context.Background(), &rpcquery.StatusParam{})
	if err != nil {
		c.Log.InfoMsg("could not get blockchain status", "err", err)
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
				c.Log.InfoMsg("Announcement", c.statusMessage()...)
			case <-doneCh:
				ticker.Stop()
				return
			}
		}
	}
}
