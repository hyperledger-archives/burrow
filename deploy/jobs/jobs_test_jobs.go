package jobs

import (
	"fmt"
	"strconv"

	hex "github.com/tmthrgd/go-hex"

	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/util"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/logging"
)

func QueryContractJob(query *def.QueryContract, do *def.DeployArgs, script *def.Playbook, client *def.Client, logger *logging.Logger) (string, []*abi.Variable, error) {
	var queryDataArray []interface{}
	var err error
	query.Function, queryDataArray, err = util.PreProcessInputData(query.Function, query.Data, do, script, client, false, logger)
	if err != nil {
		return "", nil, err
	}

	// Get the packed data from the ABI functions
	var data string
	var packedBytes []byte
	if query.Bin != "" {
		packedBytes, _, err = abi.EncodeFunctionCallFromFile(query.Bin, script.BinPath, query.Function, logger, queryDataArray...)
		data = hex.EncodeToString(packedBytes)
	}
	if query.Bin == "" || err != nil {
		packedBytes, _, err = abi.EncodeFunctionCallFromFile(query.Destination, script.BinPath, query.Function, logger, queryDataArray...)
		data = hex.EncodeToString(packedBytes)
	}
	if err != nil {
		var err = util.ABIErrorHandler(err, nil, query, logger)
		return "", nil, err
	}

	// Call the client
	txe, err := client.QueryContract(&def.QueryArg{
		Input:   query.Source,
		Address: query.Destination,
		Data:    data,
	}, logger)
	if err != nil {
		return "", nil, err
	}

	// Formally process the return
	if query.Bin != "" {
		logger.TraceMsg("Decoding Raw Result",
			"return", hex.EncodeUpperToString(txe.Result.Return),
			"Abi", query.Bin)
		query.Variables, err = abi.DecodeFunctionReturnFromFile(query.Bin, script.BinPath, query.Function, txe.Result.Return, logger)
	}
	if query.Bin == "" || err != nil {
		logger.TraceMsg("Decoding Raw Result",
			"return", hex.EncodeUpperToString(txe.Result.Return),
			"Abi", query.Destination)
		query.Variables, err = abi.DecodeFunctionReturnFromFile(query.Destination, script.BinPath, query.Function, txe.Result.Return, logger)
	}
	if err != nil {
		return "", nil, err
	}

	result2 := util.GetReturnValue(query.Variables, logger)
	// Finalize
	if result2 != "" {
		logger.InfoMsg("Return Value", "value", result2)
	} else {
		logger.InfoMsg("No return.")
	}
	return result2, query.Variables, nil
}

func QueryAccountJob(query *def.QueryAccount, client *def.Client, logger *logging.Logger) (string, error) {
	// Perform Query
	arg := fmt.Sprintf("%s:%s", query.Account, query.Field)
	logger.InfoMsg("Query Account", "argument", arg)

	result, err := util.AccountsInfo(query.Account, query.Field, client, logger)
	if err != nil {
		return "", err
	}

	// Result
	if result != "" {
		logger.InfoMsg("Return Value", "value", result)
	} else {
		logger.InfoMsg("No return.")
	}
	return result, nil
}

func QueryNameJob(query *def.QueryName, client *def.Client, logger *logging.Logger) (string, error) {
	// Peform query
	logger.InfoMsg("Querying",
		"name", query.Name,
		"field", query.Field)
	result, err := util.NamesInfo(query.Name, query.Field, client, logger)
	if err != nil {
		return "", err
	}

	if result != "" {
		logger.InfoMsg("Return Value", "result", result)
	} else {
		logger.InfoMsg("No return.")
	}
	return result, nil
}

func QueryValsJob(query *def.QueryVals, client *def.Client, logger *logging.Logger) (interface{}, error) {
	logger.InfoMsg("Querying Vals", "query", query.Query)
	result, err := util.ValidatorsInfo(query.Query, client, logger)
	if err != nil {
		return "", fmt.Errorf("error querying validators with jq-style query %s: %v", query.Query, err)
	}

	if result != nil {
		logger.InfoMsg("Return Value", "result", result)
	} else {
		logger.InfoMsg("No return.")
	}
	return result, nil
}

func AssertJob(assertion *def.Assert, logger *logging.Logger) (string, error) {
	// Switch on relation
	logger.InfoMsg("Assertion",
		"key", assertion.Key,
		"relation", assertion.Relation,
		"value", assertion.Value)

	switch assertion.Relation {
	case "==", "eq":
		/*log.Debug("Compare", strings.Compare(assertion.Key, assertion.Value))
		log.Debug("UTF8?: ", utf8.ValidString(assertion.Key))
		log.Debug("UTF8?: ", utf8.ValidString(assertion.Value))
		log.Debug("UTF8?: ", utf8.RuneCountInString(assertion.Key))
		log.Debug("UTF8?: ", utf8.RuneCountInString(assertion.Value))*/
		if assertion.Key == assertion.Value {
			return assertPass("==", assertion.Key, assertion.Value, logger)
		} else {
			return assertFail("==", assertion.Key, assertion.Value, logger)
		}
	case "!=", "ne":
		if assertion.Key != assertion.Value {
			return assertPass("!=", assertion.Key, assertion.Value, logger)
		} else {
			return assertFail("!=", assertion.Key, assertion.Value, logger)
		}
	case ">", "gt":
		k, v, err := bulkConvert(assertion.Key, assertion.Value)
		if err != nil {
			return convFail()
		}
		if k > v {
			return assertPass(">", assertion.Key, assertion.Value, logger)
		} else {
			return assertFail(">", assertion.Key, assertion.Value, logger)
		}
	case ">=", "ge":
		k, v, err := bulkConvert(assertion.Key, assertion.Value)
		if err != nil {
			return convFail()
		}
		if k >= v {
			return assertPass(">=", assertion.Key, assertion.Value, logger)
		} else {
			return assertFail(">=", assertion.Key, assertion.Value, logger)
		}
	case "<", "lt":
		k, v, err := bulkConvert(assertion.Key, assertion.Value)
		if err != nil {
			return convFail()
		}
		if k < v {
			return assertPass("<", assertion.Key, assertion.Value, logger)
		} else {
			return assertFail("<", assertion.Key, assertion.Value, logger)
		}
	case "<=", "le":
		k, v, err := bulkConvert(assertion.Key, assertion.Value)
		if err != nil {
			return convFail()
		}
		if k <= v {
			return assertPass("<=", assertion.Key, assertion.Value, logger)
		} else {
			return assertFail("<=", assertion.Key, assertion.Value, logger)
		}
	default:
		return "", fmt.Errorf("Error: Bad assert relation: \"%s\" is not a valid relation. See documentation for more information.", assertion.Relation)
	}
}

func bulkConvert(key, value string) (int, int, error) {
	k, err := strconv.Atoi(key)
	if err != nil {
		return 0, 0, err
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, 0, err
	}
	return k, v, nil
}

func assertPass(typ, key, val string, logger *logging.Logger) (string, error) {
	logger.InfoMsg("Assertion Succeeded",
		"operation", typ,
		"key", key,
		"value", val)
	return "passed", nil
}

func assertFail(typ, key, val string, logger *logging.Logger) (string, error) {
	logger.InfoMsg("Assertion Failed",
		"operation", typ,
		"key", key,
		"value", val)
	return "failed", fmt.Errorf("assertion failed")
}

func convFail() (string, error) {
	return "", fmt.Errorf("The Key of your assertion cannot be converted into an integer.\nFor string conversions please use the equal or not equal relations.")
}
