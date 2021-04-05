package service

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/vent/chain"
	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/pkg/errors"
	"github.com/tmthrgd/go-hex"
)

// buildEventData builds event data from transactions
func buildEventData(projection *sqlsol.Projection, eventClass *types.EventClass, event chain.Event,
	txOrigin *chain.Origin, evAbi *abi.EventSpec, logger *logging.Logger) (types.EventDataRow, error) {

	// a fresh new row to store column/value data
	row := make(map[string]interface{})

	// decode event data using the provided abi specification
	decodedData, err := decodeEvent(event, txOrigin, evAbi)
	if err != nil {
		return types.EventDataRow{}, errors.Wrapf(err, "Error decoding event (filter: %s)", eventClass.Filter)
	}

	logger.InfoMsg("Decoded event", decodedData)

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
			if bs, ok := value.(*[]byte); ok {
				if fieldMapping.BytesToString {
					str := sanitiseBytesForString(*bs, logger)
					value = interface{}(str)
				} else if fieldMapping.BytesToHex {
					value = hex.EncodeUpperToString(*bs)
				}
			}
			row[column.Name] = value
		} else {
			logger.TraceMsg("could not get column", "err", err)
		}
	}

	return types.EventDataRow{Action: rowAction, RowData: row, EventClass: eventClass}, nil
}

func buildBlkData(tbls types.EventTables, block chain.Block) (types.EventDataRow, error) {
	// block raw data
	if _, ok := tbls[tables.Block]; ok {
		row, err := block.GetMetadata(columns)
		if err != nil {
			return types.EventDataRow{}, err
		}
		return types.EventDataRow{Action: types.ActionUpsert, RowData: row}, nil
	}
	return types.EventDataRow{}, fmt.Errorf("table: %s not found in table structure %v", tables.Block, tbls)

}

// buildTxData builds transaction data from tx stream
func buildTxData(txe chain.Transaction) (types.EventDataRow, error) {
	row, err := txe.GetMetadata(columns)
	if err != nil {
		return types.EventDataRow{}, fmt.Errorf("could not get transaction metadata: %w", err)
	}

	return types.EventDataRow{
		Action:  types.ActionUpsert,
		RowData: row,
	}, nil
}

func sanitiseBytesForString(bs []byte, l *logging.Logger) string {
	str, err := UTF8StringFromBytes(bs)
	if err != nil {
		l.InfoMsg("buildEventData() received invalid bytes for utf8 string - proceeding with sanitised version",
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
