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
func buildEventData(projection *sqlsol.Projection, eventClass *types.EventClass, event *exec.Event, origin *exec.Origin, abiSpec *abi.AbiSpec,
	l *logger.Logger) (types.EventDataRow, error) {

	// a fresh new row to store column/value data
	row := make(map[string]interface{})

	// get header & log data for the given event
	eventHeader := event.GetHeader()
	eventLog := event.GetLog()

	// decode event data using the provided abi specification
	decodedData, err := decodeEvent(eventHeader, eventLog, origin, abiSpec)
	if err != nil {
		return types.EventDataRow{}, errors.Wrapf(err, "Error decoding event (filter: %s)", eventClass.Filter)
	}

	l.Info("msg", fmt.Sprintf("Unpacked data: %v", decodedData), "eventName", decodedData[types.EventNameLabel])

	rowAction := types.ActionUpsert

	// for each data element, maps to SQL columnName and gets its value
	// if there is no matching column for the item, it doesn't need to be stored in db
	for fieldName, value := range decodedData {
		// Can't think of case where we will get a key that is empty, but if we ever did we should not treat
		// it as a delete marker when the delete marker field in unset
		if eventClass.DeleteMarkerField != "" && eventClass.DeleteMarkerField == fieldName {
			rowAction = types.ActionDelete
		}
		fieldMapping := eventClass.GetFieldMapping(fieldName)
		if fieldMapping == nil {
			continue
		}
		column, err := projection.GetColumn(eventClass.TableName, fieldMapping.ColumnName)
		if err == nil {
			if fieldMapping.BytesToString {
				if bs, ok := value.(*[]byte); ok {
					str := sanitiseBytesForString(*bs, l)
					row[column.Name] = interface{}(str)
					continue
				}
			}
			row[column.Name] = value
		} else {
			l.Debug("msg", "could not get column", "err", err)
		}
	}

	return types.EventDataRow{Action: rowAction, RowData: row, EventClass: eventClass}, nil
}

// buildBlkData builds block data from block stream
func buildBlkData(tbls types.EventTables, block *exec.BlockExecution) (types.EventDataRow, error) {
	// a fresh new row to store column/value data
	row := make(map[string]interface{})

	// block raw data
	if _, ok := tbls[tables.Block]; ok {
		blockHeader, err := json.Marshal(block.Header)
		if err != nil {
			return types.EventDataRow{}, fmt.Errorf("couldn not marshal BlockHeader in block %v", block)
		}

		row[columns.Height] = fmt.Sprintf("%v", block.Height)
		row[columns.BlockHeader] = string(blockHeader)
	} else {
		return types.EventDataRow{}, fmt.Errorf("table: %s not found in table structure %v", tables.Block, tbls)
	}

	return types.EventDataRow{Action: types.ActionUpsert, RowData: row}, nil
}

// buildTxData builds transaction data from tx stream
func buildTxData(txe *exec.TxExecution) (types.EventDataRow, error) {
	// transaction raw data
	envelope, err := json.Marshal(txe.Envelope)
	if err != nil {
		return types.EventDataRow{}, fmt.Errorf("couldn't marshal envelope in tx %v: %v", txe, err)
	}

	events, err := json.Marshal(txe.Events)
	if err != nil {
		return types.EventDataRow{}, fmt.Errorf("couldn't marshal events in tx %v: %v", txe, err)
	}

	result, err := json.Marshal(txe.Result)
	if err != nil {
		return types.EventDataRow{}, fmt.Errorf("couldn't marshal result in tx %v: %v", txe, err)
	}

	receipt, err := json.Marshal(txe.Receipt)
	if err != nil {
		return types.EventDataRow{}, fmt.Errorf("couldn't marshal receipt in tx %v: %v", txe, err)
	}

	exception, err := json.Marshal(txe.Exception)
	if err != nil {
		return types.EventDataRow{}, fmt.Errorf("couldn't marshal exception in tx %v: %v", txe, err)
	}

	origin, err := json.Marshal(txe.Origin)
	if err != nil {
		return types.EventDataRow{}, fmt.Errorf("couldn't marshal origin in tx %v: %v", txe, err)
	}

	return types.EventDataRow{
		Action: types.ActionUpsert,
		RowData: map[string]interface{}{
			columns.Height:    txe.Height,
			columns.TxHash:    txe.TxHash.String(),
			columns.Index:     txe.Index,
			columns.TxType:    txe.TxType.String(),
			columns.Envelope:  string(envelope),
			columns.Events:    string(events),
			columns.Result:    string(result),
			columns.Receipt:   string(receipt),
			columns.Origin:    string(origin),
			columns.Exception: string(exception),
		},
	}, nil
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
