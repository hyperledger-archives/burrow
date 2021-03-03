package service

import (
	"io"

	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/vent/chain"
	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/pkg/errors"
)

func NewBlockConsumer(chainID string, projection *sqlsol.Projection, opt sqlsol.SpecOpt, getEventSpec EventSpecGetter,
	eventCh chan<- types.EventData, doneCh chan struct{}, logger *logging.Logger) func(block chain.Block) error {

	logger = logger.WithScope("makeBlockConsumer")

	var blockHeight uint64

	return func(block chain.Block) error {
		if finished(doneCh) {
			return io.EOF
		}

		// set new block number
		blockHeight = block.GetHeight()
		txs := block.GetTxs()

		logger.TraceMsg("Block received",
			"height", blockHeight,
			"num_txs", len(txs))

		// create a fresh new structure to store block data at this height
		blockData := sqlsol.NewBlockData(blockHeight)

		if opt.Enabled(sqlsol.Block) {
			blkRawData, err := buildBlkData(projection.Tables, block)
			if err != nil {
				return errors.Wrapf(err, "Error building block raw data")
			}
			// set row in structure
			blockData.AddRow(tables.Block, blkRawData)
		}

		for _, txe := range txs {
			events := txe.GetEvents()
			logger.TraceMsg("Getting transaction", "TxHash", txe.GetHash(), "num_events", len(events))

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
			if txe.GetException() == nil {
				txOrigin := txe.GetOrigin()
				if txOrigin == nil {
					// This is an original transaction from the current chain so we build its origin from context
					txOrigin = &chain.Origin{
						ChainID: chainID,
						Height:  block.GetHeight(),
						Index:   txe.GetIndex(),
					}
				}

				for _, event := range events {
					var tagged query.Tagged = event
					eventID := exec.SolidityEventID(event.GetTopics())
					eventSpec, eventSpecErr := getEventSpec(eventID, event.GetAddress())
					if eventSpecErr != nil {
						logger.InfoMsg("could not get ABI for solidity event",
							structure.ErrorKey, eventSpecErr,
							"event_id", eventID,
							"address", event.GetAddress())
					} else {
						// Since we have the event ABI we will allow matching on ABI fields
						tagged = query.TagsFor(event, query.TaggedPrefix("Event", eventSpec))
					}

					// see which spec filter matches with the one in event data
					for _, eventClass := range projection.Spec {
						qry, err := eventClass.Query()

						if err != nil {
							return errors.Wrapf(err, "Error parsing query from filter string")
						}

						// there's a matching filter, add data to the rows
						if qry.Matches(tagged) {
							if eventSpecErr != nil {
								return errors.Wrapf(eventSpecErr, "could not get ABI for solidity event matching "+
									"projection filter \"%s\" with id %v at address %v",
									eventClass.Filter, eventID, event.GetAddress())
							}

							logger.InfoMsg("Matched event", "event_id", eventID, "filter", eventClass.Filter)

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
		for name, rows := range blockData.Data.Tables {
			logger.InfoMsg("Upserting rows in SQL table", "height", blockHeight, "table", name, "action", "UPSERT", "rows", rows)
		}

		eventCh <- blockData.Data
		return nil
	}
}
