package service

import (
	"math/big"
	"strconv"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/pkg/errors"
)

// decodeEvent unpacks & decodes event data
func decodeEvent(eventHeader *exec.Header, log *exec.LogEvent, txOrigin *exec.Origin, evAbi *abi.EventSpec) (map[string]interface{}, error) {
	// to prepare decoded data and map to event item name
	data := make(map[string]interface{})

	// decode header to get context data for each event
	data[types.EventNameLabel] = evAbi.Name
	data[types.ChainIDLabel] = txOrigin.ChainID
	data[types.BlockHeightLabel] = strconv.FormatUint(txOrigin.GetHeight(), 10)
	data[types.TxIndexLabel] = strconv.FormatUint(txOrigin.GetIndex(), 10)
	data[types.EventIndexLabel] = strconv.FormatUint(eventHeader.GetIndex(), 10)
	data[types.EventTypeLabel] = eventHeader.GetEventType().String()
	data[types.TxTxHashLabel] = eventHeader.TxHash.String()

	// build expected interface type array to get log event values
	unpackedData := abi.GetPackingTypes(evAbi.Inputs)

	// unpack event data (topics & data part)
	if err := abi.UnpackEvent(evAbi, log.Topics, log.Data, unpackedData...); err != nil {
		return nil, errors.Wrap(err, "Could not unpack event data")
	}

	// for each decoded item value, stores it in given item name
	for i, input := range evAbi.Inputs {
		switch v := unpackedData[i].(type) {
		case *crypto.Address:
			data[input.Name] = v.String()
		case *big.Int:
			data[input.Name] = v.String()
		case *string:
			data[input.Name] = *v
		default:
			data[input.Name] = v
		}
	}

	return data, nil
}
