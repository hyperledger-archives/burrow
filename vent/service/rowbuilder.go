package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/vent/logger"
	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/pkg/errors"
)

// buildEventData builds event data from transactions
func buildEventData(spec types.EventDefinition, parser *sqlsol.Parser, event *exec.Event, abiSpec *abi.AbiSpec, l *logger.Logger) (types.EventDataRow, error) {

	// a fresh new row to store column/value data
	row := make(map[string]interface{})

	// a replacer to get DeleteFilter parameters
	replacer := strings.NewReplacer(" ", "", "'", "")

	// get header & log data for the given event
	eventHeader := event.GetHeader()
	eventLog := event.GetLog()

	// decode event data using the provided abi specification
	decodedData, err := decodeEvent(eventHeader, eventLog, abiSpec)
	if err != nil {
		return types.EventDataRow{}, errors.Wrapf(err, "Error decoding event (filter: %s)", spec.Filter)
	}

	l.Info("msg", fmt.Sprintf("Unpacked data: %v", decodedData), "eventName", decodedData[types.EventNameLabel])

	rowAction := types.ActionUpsert

	var deleteFilter []string

	// get delete filter from spec
	if spec.DeleteFilter != "" {
		deleteFilter = strings.Split(replacer.Replace(spec.DeleteFilter), "=")
	}
	deleteFilterLength := len(deleteFilter)

	// for each data element, maps to SQL columnName and gets its value
	// if there is no matching column for the item, it doesn't need to be stored in db
	for k, v := range decodedData {
		if deleteFilterLength > 0 {
			if k == deleteFilter[0] {
				if bs, ok := v.(*[]byte); ok {
					str := sanitiseBytesForString(*bs, l)
					if str == deleteFilter[1] {
						rowAction = types.ActionDelete
					}
				}
			}
		}
		if column, err := parser.GetColumn(spec.TableName, k); err == nil {
			if column.BytesToString {
				if bs, ok := v.(*[]byte); ok {
					str := sanitiseBytesForString(*bs, l)
					row[column.Name] = interface{}(str)
					continue
				}
			}
			row[column.Name] = v
		}
	}

	return types.EventDataRow{Action: rowAction, RowData: row}, nil
}

// buildBlkData builds block data from block stream
func buildBlkData(tbls types.EventTables, block *exec.BlockExecution) (types.EventDataRow, error) {
	// a fresh new row to store column/value data
	row := make(map[string]interface{})

	// block raw data
	if tbl, ok := tbls[types.SQLBlockTableName]; ok {

		blockHeader, err := json.Marshal(block.BlockHeader)
		if err != nil {
			return types.EventDataRow{}, fmt.Errorf("Couldn't marshal BlockHeader in block %v", block)
		}

		row[tbl.Columns[types.BlockHeightLabel].Name] = fmt.Sprintf("%v", block.Height)
		row[tbl.Columns[types.BlockHeaderLabel].Name] = string(blockHeader)
	} else {
		return types.EventDataRow{}, fmt.Errorf("table: %s not found in table structure %v", types.SQLBlockTableName, tbls)
	}

	return types.EventDataRow{Action: types.ActionUpsert, RowData: row}, nil
}

// buildTxData builds transaction data from tx stream
func buildTxData(tbls types.EventTables, txe *exec.TxExecution) (types.EventDataRow, error) {

	// a fresh new row to store column/value data
	row := make(map[string]interface{})

	// transaction raw data
	if tbl, ok := tbls[types.SQLTxTableName]; ok {

		envelope, err := json.Marshal(txe.Envelope)
		if err != nil {
			return types.EventDataRow{}, fmt.Errorf("Couldn't marshal envelope in tx %v", txe)
		}

		events, err := json.Marshal(txe.Events)
		if err != nil {
			return types.EventDataRow{}, fmt.Errorf("Couldn't marshal events in tx %v", txe)
		}

		result, err := json.Marshal(txe.Result)
		if err != nil {
			return types.EventDataRow{}, fmt.Errorf("Couldn't marshal result in tx %v", txe)
		}

		receipt, err := json.Marshal(txe.Receipt)
		if err != nil {
			return types.EventDataRow{}, fmt.Errorf("Couldn't marshal receipt in tx %v", txe)
		}

		exception, err := json.Marshal(txe.Exception)
		if err != nil {
			return types.EventDataRow{}, fmt.Errorf("Couldn't marshal exception in tx %v", txe)
		}

		row[tbl.Columns[types.BlockHeightLabel].Name] = fmt.Sprintf("%v", txe.Height)
		row[tbl.Columns[types.TxTxHashLabel].Name] = txe.TxHash.String()
		row[tbl.Columns[types.TxIndexLabel].Name] = txe.Index
		row[tbl.Columns[types.TxTxTypeLabel].Name] = txe.TxType.String()
		row[tbl.Columns[types.TxEnvelopeLabel].Name] = string(envelope)
		row[tbl.Columns[types.TxEventsLabel].Name] = string(events)
		row[tbl.Columns[types.TxResultLabel].Name] = string(result)
		row[tbl.Columns[types.TxReceiptLabel].Name] = string(receipt)
		row[tbl.Columns[types.TxExceptionLabel].Name] = string(exception)
	} else {
		return types.EventDataRow{}, fmt.Errorf("Table: %s not found in table structure %v", types.SQLTxTableName, tbls)
	}

	return types.EventDataRow{Action: types.ActionUpsert, RowData: row}, nil
}

func sanitiseBytesForString(bs []byte, l *logger.Logger) string {
	str, err := UTF8StringFromBytes(bs)
	if err != nil {
		l.Error("msg", "buildEventData() received invalid bytes for utf8 string - proceeding with sanitised version",
			"err", err)
	}
	// The only null bytes in utf8 are for the null code point/character so this is fine in general
	return strings.Trim(str, "\x00")
}

// Checks whether the bytes passed are valid utf8 string bytes. If they are not returns a sanitised string version of the
// bytes with offending sequences replaced by the utf8 replacement/error rune and an error indicating the offending
// byte sequences and their position. Note: always returns a valid string regardless of error.
func UTF8StringFromBytes(bs []byte) (string, error) {
	// Provide fast path for good strings
	if utf8.Valid(bs) {
		return string(bs), nil
	}
	buf := new(bytes.Buffer)
	var runeErrs []string
	// This loops over runs (code points and unlike range of string gives us index of code point (i.e. utf8 char)
	// not bytes, which we want for error message
	var offset int
	// Iterate over character indices (not byte indices)
	for i := 0; i < len(bs); i++ {
		r, n := utf8.DecodeRune(bs[offset:])
		buf.WriteRune(r)
		if r == utf8.RuneError {
			runeErrs = append(runeErrs, fmt.Sprintf("0x% X (at index %d)", bs[offset:offset+n], i))
		}
		offset += n
	}
	str := buf.String()
	errHeader := fmt.Sprintf("bytes purported to represent the string '%s'", str)
	switch len(runeErrs) {
	case 0:
		// should not happen
		return str, fmt.Errorf("bytes appear to be invalid utf8 but do not contain invalid code points")
	case 1:
		return str, fmt.Errorf("%s contain invalid utf8 byte sequence: %s", errHeader, runeErrs[0])
	default:
		return str, fmt.Errorf("%s contain invalid utf8 byte sequences: %s", errHeader,
			strings.Join(runeErrs, ", "))
	}
}
