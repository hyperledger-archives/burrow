package service

import (
	"math/big"
	"strconv"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/vent/chain"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/pkg/errors"
)

// decodeEvent unpacks & decodes event data
func decodeEvent(log chain.Event, txOrigin *chain.Origin, evAbi *abi.EventSpec) (map[string]interface{}, error) {
	// to prepare decoded data and map to event item name
	data := make(map[string]interface{})

	// decode header to get context data for each event
	data[types.EventNameLabel] = evAbi.Name
	data[types.ChainIDLabel] = txOrigin.ChainID
	data[types.BlockHeightLabel] = strconv.FormatUint(txOrigin.Height, 10)
	data[types.TxIndexLabel] = strconv.FormatUint(txOrigin.Index, 10)
	data[types.EventIndexLabel] = strconv.FormatUint(log.GetIndex(), 10)
	data[types.EventTypeLabel] = exec.TypeLog.String()
	data[types.TxTxHashLabel] = log.GetTransactionHash().String()

	// build expected interface type array to get log event values
	unpackedData := abi.GetPackingTypes(evAbi.Inputs)

	// unpack event data (topics & data part)
	if err := abi.UnpackEvent(evAbi, log.GetTopics(), log.GetData(), unpackedData...); err != nil {
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
