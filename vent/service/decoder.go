package service

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/pkg/errors"
)

// decodeEvent unpacks & decodes event data
func decodeEvent(header *exec.Header, log *exec.LogEvent, origin *exec.Origin, abiSpec *abi.Spec) (map[string]interface{}, error) {
	// to prepare decoded data and map to event item name
	data := make(map[string]interface{})

	var eventID abi.EventID
	copy(eventID[:], log.Topics[0].Bytes())

	evAbi, ok := abiSpec.EventsById[eventID]
	if !ok {
		return nil, fmt.Errorf("abi spec not found for event %x", eventID)
	}

	// decode header to get context data for each event
	data[types.EventNameLabel] = evAbi.Name
	data[types.ChainIDLabel] = origin.ChainID
	data[types.BlockHeightLabel] = fmt.Sprintf("%v", origin.GetHeight())
	data[types.BlockIndexLabel] = fmt.Sprintf("%v", origin.GetIndex())
	data[types.EventTypeLabel] = header.GetEventType().String()
	data[types.TxTxHashLabel] = header.TxHash.String()

	// build expected interface type array to get log event values
	unpackedData := abi.GetPackingTypes(evAbi.Inputs)

	// unpack event data (topics & data part)
	if err := abi.UnpackEvent(&evAbi, log.Topics, log.Data, unpackedData...); err != nil {
		return nil, errors.Wrap(err, "could not unpack event data")
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
