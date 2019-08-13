package service

import (
	"io"
	"reflect"

	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/pkg/errors"
)

func NewBlockConsumer(projection *sqlsol.Projection, opt sqlsol.SpecOpt, getEventSpec EventSpecGetter,
	eventCh chan<- types.EventData, doneCh chan struct{},
	logger *logging.Logger) func(blockExecution *exec.BlockExecution) error {

	logger = logger.WithScope("makeBlockConsumer")

	return func(blockExecution *exec.BlockExecution) error {
		if finished(doneCh) {
			return io.EOF
		}

		// set new block number
		fromBlock := blockExecution.Height

		logger.TraceMsg("Block received",
			"height", blockExecution.Height,
			"num_txs", len(blockExecution.TxExecutions))

		// create a fresh new structure to store block data at this height
		blockData := sqlsol.NewBlockData(fromBlock)

		if opt.Enabled(sqlsol.Block) {
			blkRawData, err := buildBlkData(projection.Tables, blockExecution)
			if err != nil {
				return errors.Wrapf(err, "Error building block raw data")
			}
			// set row in structure
			blockData.AddRow(tables.Block, blkRawData)
		}

		// get transactions for a given block
		for _, txe := range blockExecution.TxExecutions {
			logger.TraceMsg("Getting transaction", "TxHash", txe.TxHash, "num_events", len(txe.Events))

			if opt.Enabled(sqlsol.Tx) {
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
				txOrigin := txe.Origin
				if txOrigin == nil {
					// This is an original transaction from the current chain so we build its origin from context
					txOrigin = &exec.Origin{
						Time:    blockExecution.GetHeader().GetTime(),
						ChainID: blockExecution.GetHeader().GetChainID(),
						Height:  txe.GetHeight(),
						Index:   txe.GetIndex(),
					}
				}

				// get events for a given transaction
				for _, event := range txe.Events {
					if event.Log == nil {
						// Only EVM events are of interest
						continue
					}

					eventID := event.Log.SolidityEventID()
					eventSpec, err := getEventSpec(eventID, event.Log.Address)
					if err != nil {
						return errors.Wrapf(err, "could not get ABI for solidity event with id %v at address %v",
							eventID, event.Log.Address)
					}
					tagged := query.TagsFor(event, query.TaggedPrefix("Event", eventSpec))

					// see which spec filter matches with the one in event data
					for _, eventClass := range projection.Spec {
						qry, err := eventClass.Query()

						if err != nil {
							return errors.Wrapf(err, "Error parsing query from filter string")
						}

						// there's a matching filter, add data to the rows
						if qry.Matches(tagged) {

							logger.InfoMsg("Matched event", "header", event.Header,
								"filter", eventClass.Filter)

							// unpack, decode & build event data
							eventData, err := buildEventData(projection, eventClass, event, txOrigin, eventSpec, logger)
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
				logger.InfoMsg("Upserting rows in SQL table", "height", fromBlock, "table", name, "action", "UPSERT", "rows", rows)
			}

			eventCh <- blk
		}
		return nil
	}
}

type eventSpecTagged struct {
	Event abi.EventSpec
}

func (e *eventSpecTagged) Get(key string) (value interface{}, ok bool) {
	return query.GetReflect(reflect.ValueOf(e), key)
}
