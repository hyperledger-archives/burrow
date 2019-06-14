package jobs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"

	"github.com/hyperledger/burrow/crypto"
	compilers "github.com/hyperledger/burrow/deploy/compile"
	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/util"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/txs/payload"
	hex "github.com/tmthrgd/go-hex"
)

var errCodeMissing = fmt.Errorf("error: no binary code found in contract. Contract may be abstract due to missing function body or inherited function signatures not matching.")

func BuildJob(build *def.Build, deployScript *def.Playbook, resp *compilers.Response, logger *logging.Logger) (result string, err error) {
	// assemble contract
	contractPath, err := findContractFile(build.Contract, deployScript.BinPath, deployScript.Path)
	if err != nil {
		return
	}

	logger.InfoMsg("Contract path", "path", contractPath)

	// normal compilation/deploy sequence
	if resp == nil {
		logger.InfoMsg("Error compiling contracts: Missing compiler result")
		return "", fmt.Errorf("internal error")
	} else if resp.Error != "" {
		logger.InfoMsg("Error compiling contracts", "Language error", resp.Error)
		return "", fmt.Errorf("%v", resp.Error)
	} else if resp.Warning != "" {
		logger.InfoMsg("Warning during contraction compilation", "warning", resp.Warning)
	}

	// Save
	binP := build.BinPath
	if binP == "" {
		binP = deployScript.BinPath
	} else {
		if _, err := os.Stat(binP); os.IsNotExist(err) {
			if err := os.Mkdir(binP, 0775); err != nil {
				return "", err
			}
		}
	}

	for _, res := range resp.Objects {
		switch build.Instance {
		case "":
			if res.Filename != build.Contract {
				logger.TraceMsg("Ignoring output for different solidity file", "found", res.Filename, "expected", build.Contract)
				continue
			}
		case "all":
		default:
			if res.Objectname != build.Instance {
				continue
			}
		}

		// saving binary
		logger.InfoMsg("Saving Binary", "name", res.Objectname, "dir", binP)

		err = res.Contract.Save(binP, fmt.Sprintf("%s.bin", res.Objectname))
		if err != nil {
			return "", err
		}

		if build.Store != "" {
			dir := filepath.Dir(build.Store)
			file := filepath.Base(build.Store)

			if _, err := os.Stat(dir); os.IsNotExist(err) {
				if err := os.Mkdir(dir, 0775); err != nil {
					return "", err
				}
			}

			err = res.Contract.Save(dir, file)
			if err != nil {
				return "", err
			}
		}
	}

	return "", nil
}

func FormulateDeployJob(deploy *def.Deploy, do *def.DeployArgs, deployScript *def.Playbook, client *def.Client, intermediate interface{}, logger *logging.Logger) (txs []*payload.CallTx, contracts []*compilers.ResponseItem, err error) {
	deploy.Libraries, _ = util.PreProcessLibs(deploy.Libraries, do, deployScript, client, logger)
	// trim the extension and path
	contractName := filepath.Base(deploy.Contract)
	contractName = strings.TrimSuffix(contractName, filepath.Ext(contractName))

	// Use defaults
	deploy.Source = FirstOf(deploy.Source, deployScript.Account)
	deploy.Instance = FirstOf(deploy.Instance, contractName)
	deploy.Amount = FirstOf(deploy.Amount, do.DefaultAmount)
	deploy.Fee = FirstOf(deploy.Fee, do.DefaultFee)
	deploy.Gas = FirstOf(deploy.Gas, do.DefaultGas)

	// assemble contract
	contractPath, err := findContractFile(deploy.Contract, deployScript.BinPath, deployScript.Path)
	if err != nil {
		return
	}

	txs = make([]*payload.CallTx, 0)
	libs := make(map[string]string)
	var list []string
	if strings.Contains(deploy.Libraries, " ") {
		list = strings.Split(deploy.Libraries, " ")
	} else {
		list = strings.Split(deploy.Libraries, ",")
	}
	for _, l := range list {
		if l != "" {
			v := strings.Split(l, ":")
			if len(v) != 2 {
				return nil, nil, fmt.Errorf("library %s should be contract:address format", l)
			}
			libs[v[0]] = v[1]
		}
	}

	contracts = make([]*compilers.ResponseItem, 0)

	// compile
	if filepath.Ext(deploy.Contract) != ".sol" {
		logger.InfoMsg("Binary file detected. Using binary deploy sequence.", "Binary path", contractPath)

		contract, err := compilers.LoadSolidityContract(contractPath)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to read contract %s: %v", contractPath, err)
		}
		err = contract.Link(libs)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to link contract %s: %v", contractPath, err)
		}
		contractCode := contract.Evm.Bytecode.Object

		mergeAbiSpecBytes(client, contract.Abi)

		if deploy.Data != nil {
			_, callDataArray, err := util.PreProcessInputData("", deploy.Data, do, deployScript, client, true, logger)
			if err != nil {
				return nil, nil, err
			}
			packedBytes, _, err := abi.EncodeFunctionCall(string(contract.Abi), "", logger, callDataArray...)
			if err != nil {
				return nil, nil, err
			}
			callData := hex.EncodeToString(packedBytes)
			contractCode = contractCode + callData
		}

		tx, err := deployTx(client, deploy, contractName, string(contractCode), logger)
		if err != nil {
			return nil, nil, fmt.Errorf("could not deploy binary contract: %v", err)
		}
		txs = []*payload.CallTx{tx}
		contracts = append(contracts, &compilers.ResponseItem{Filename: contractPath, Objectname: contractName, Contract: *contract})
	} else {
		contractPath = deploy.Contract
		logger.InfoMsg("Contract path", "path", contractPath)
		// normal compilation/deploy sequence

		resp, err := getCompilerWork(intermediate)
		if err != nil {
			return nil, nil, err
		}

		if resp == nil {
			logger.InfoMsg("Error compiling contracts: Missing compiler result")
			return nil, nil, fmt.Errorf("internal error")
		} else if resp.Error != "" {
			logger.InfoMsg("Error compiling contracts: Language error:", "error", resp.Error)
			return nil, nil, fmt.Errorf("%v", resp.Error)
		} else if resp.Warning != "" {
			logger.InfoMsg("Warning during contract compilation", "warning", resp.Warning)
		}
		// loop through objects returned from compiler
		switch {
		case len(resp.Objects) == 1:
			response := resp.Objects[0]
			logger.TraceMsg("Deploying the single contract from solidity file",
				"path", contractPath,
				"abi", string(response.Contract.Abi),
				"bin", response.Contract.Evm.Bytecode.Object)
			if response.Contract.Evm.Bytecode.Object == "" {
				return nil, nil, errCodeMissing
			}
			mergeAbiSpecBytes(client, response.Contract.Abi)

			tx, err := deployContract(deploy, do, deployScript, client, response, libs, logger)
			if err != nil {
				return nil, nil, err
			}

			txs = []*payload.CallTx{tx}
			contracts = append(contracts, &resp.Objects[0])
		case deploy.Instance == "all":
			logger.InfoMsg("Deploying all contracts", "path", contractPath)
			var baseObj *payload.CallTx
			var baseContract *compilers.ResponseItem
			deployedCount := 0
			for i, response := range resp.Objects {
				if response.Contract.Evm.Bytecode.Object == "" {
					continue
				}
				mergeAbiSpecBytes(client, response.Contract.Abi)
				tx, err := deployContract(deploy, do, deployScript, client, response, libs, logger)
				if err != nil {
					return nil, nil, err
				}
				deployedCount++
				if strings.ToLower(response.Objectname) == strings.ToLower(strings.TrimSuffix(filepath.Base(deploy.Contract), filepath.Ext(filepath.Base(deploy.Contract)))) {
					baseObj = tx
					baseContract = &resp.Objects[i]
				} else {
					txs = append(txs, tx)
					contracts = append(contracts, &resp.Objects[i])
				}
			}

			// Make sure the Contact which matches the filename is last, so that addres is used
			if baseObj != nil {
				txs = append(txs, baseObj)
				contracts = append(contracts, baseContract)
			} else if deployedCount == 0 {
				return nil, nil, errCodeMissing
			}

		default:
			logger.InfoMsg("Deploying a single contract that matches", "contract", deploy.Instance)
			for i, response := range resp.Objects {
				if response.Contract.Evm.Bytecode.Object == "" ||
					response.Filename != deploy.Contract {
					continue
				}
				if matchInstanceName(response.Objectname, deploy.Instance) {
					if response.Contract.Evm.Bytecode.Object == "" {
						return nil, nil, errCodeMissing
					}
					logger.TraceMsg("Deploy contract",
						"contract", response.Objectname,
						"Abi", string(response.Contract.Abi),
						"Bin", response.Contract.Evm.Bytecode.Object)
					tx, err := deployContract(deploy, do, deployScript, client, response, libs, logger)
					if err != nil {
						return nil, nil, err
					}
					txs = append(txs, tx)
					// make sure we copy response, as it is the loop variable and will be overwritten
					contracts = append(contracts, &resp.Objects[i])
				}
			}
		}
	}

	return
}

func DeployJob(deploy *def.Deploy, do *def.DeployArgs, script *def.Playbook, client *def.Client, txs []*payload.CallTx, contracts []*compilers.ResponseItem, logger *logging.Logger) (result string, err error) {
	// saving contract
	// additional data may be sent along with the contract
	// these are naively added to the end of the contract code using standard
	// mint packing

	for i, tx := range txs {
		// Sign, broadcast, display
		contractAddress, err := deployFinalize(do, client, tx, logger)
		if err != nil {
			return "", fmt.Errorf("Error finalizing contract deploy %s: %v", deploy.Contract, err)
		}

		// saving contract/library abi at abi/address
		if contracts != nil && contractAddress != nil {
			contract := contracts[i].Contract
			// saving binary
			logger.TraceMsg("Saving Binary", "address", contractAddress.String())
			err = contract.Save(script.BinPath, fmt.Sprintf("%s.bin", contractAddress.String()))
			if err != nil {
				return "", err
			}
			result = contractAddress.String()
		} else {
			// we shouldn't reach this point because we should have an error before this.
			return "", fmt.Errorf("The contract did not deploy. Unable to save abi to abi/contractAddress.")
		}
	}

	return result, nil
}

func matchInstanceName(objectName, deployInstance string) bool {
	if objectName == "" {
		return false
	}
	// Ignore the filename component that newer versions of Solidity include in object name

	objectNameParts := strings.Split(objectName, ":")
	deployInstanceParts := strings.Split(deployInstance, "/")
	return strings.ToLower(objectNameParts[len(objectNameParts)-1]) == strings.ToLower(deployInstanceParts[len(deployInstanceParts)-1])
}

func findContractFile(contract, binPath string, deployPath string) (string, error) {
	contractPaths := []string{
		contract,
		filepath.Join(binPath, contract),
		filepath.Join(binPath, filepath.Base(contract)),
		filepath.Join(deployPath, contract),
		filepath.Join(deployPath, filepath.Base(contract)),
	}

	for _, p := range contractPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("Could not find contract in any of %v", contractPaths)
}

// TODO [rj] refactor to remove [contractPath] from functions signature => only used in a single error throw.
func deployContract(deploy *def.Deploy, do *def.DeployArgs, script *def.Playbook, client *def.Client, compilersResponse compilers.ResponseItem, libs map[string]string, logger *logging.Logger) (*payload.CallTx, error) {
	contract := compilersResponse.Contract
	contractName := compilersResponse.Objectname
	logger.InfoMsg("Saving Binary", "contract", contractName)
	err := contract.Save(script.BinPath, fmt.Sprintf("%s.bin", contractName))
	if err != nil {
		return nil, err
	}

	if deploy.Store != "" {
		dir := filepath.Dir(deploy.Store)
		file := filepath.Base(deploy.Store)

		err = contract.Save(dir, file)
		if err != nil {
			return nil, err
		}
	}

	err = contract.Link(libs)
	if err != nil {
		return nil, err
	}
	contractCode := contract.Evm.Bytecode.Object

	if deploy.Data != nil {
		_, callDataArray, err := util.PreProcessInputData(compilersResponse.Objectname, deploy.Data, do, script, client, true, logger)
		if err != nil {
			return nil, err
		}
		packedBytes, _, err := abi.EncodeFunctionCall(string(compilersResponse.Contract.Abi), "", logger, callDataArray...)
		if err != nil {
			return nil, err
		}
		callData := hex.EncodeToString(packedBytes)
		contractCode = contractCode + callData
	} else {
		// No constructor arguments were provided. Did the constructor want any?
		spec, err := abi.ReadAbiSpec(compilersResponse.Contract.Abi)
		if err != nil {
			return nil, err
		}

		if len(spec.Constructor.Inputs) > 0 {
			logger.InfoMsg("Constructor wants %d arguments but 0 provided", len(spec.Constructor.Inputs))
			return nil, fmt.Errorf("Constructor wants %d arguments but 0 provided", len(spec.Constructor.Inputs))
		}
	}

	return deployTx(client, deploy, compilersResponse.Objectname, contractCode, logger)
}

func deployTx(client *def.Client, deploy *def.Deploy, contractName, contractCode string, logger *logging.Logger) (*payload.CallTx, error) {
	// Deploy contract
	logger.TraceMsg("Deploying Contract",
		"contract", contractName,
		"source", deploy.Source,
		"code", contractCode,
		"chain", client.ChainAddress)

	return client.Call(&def.CallArg{
		Input:    deploy.Source,
		Amount:   deploy.Amount,
		Fee:      deploy.Fee,
		Gas:      deploy.Gas,
		Data:     contractCode,
		Sequence: deploy.Sequence,
	}, logger)
}

func FormulateCallJob(call *def.Call, do *def.DeployArgs, deployScript *def.Playbook, client *def.Client, logger *logging.Logger) (tx *payload.CallTx, err error) {
	var callData string
	var callDataArray []interface{}
	//todo: find a way to call the fallback function here
	call.Function, callDataArray, err = util.PreProcessInputData(call.Function, call.Data, do, deployScript, client, false, logger)
	if err != nil {
		return nil, err
	}
	// Use default
	call.Source = FirstOf(call.Source, deployScript.Account)
	call.Amount = FirstOf(call.Amount, do.DefaultAmount)
	call.Fee = FirstOf(call.Fee, do.DefaultFee)
	call.Gas = FirstOf(call.Gas, do.DefaultGas)

	// formulate call
	var packedBytes []byte
	var funcSpec *abi.FunctionSpec
	logger.TraceMsg("Looking for ABI in", "path", deployScript.BinPath, "bin", call.Bin, "dest", call.Destination)
	if call.Bin != "" {
		packedBytes, funcSpec, err = abi.EncodeFunctionCallFromFile(call.Bin, deployScript.BinPath, call.Function, logger, callDataArray...)
		callData = hex.EncodeToString(packedBytes)
	}
	if call.Bin == "" || err != nil {
		packedBytes, funcSpec, err = abi.EncodeFunctionCallFromFile(call.Destination, deployScript.BinPath, call.Function, logger, callDataArray...)
		callData = hex.EncodeToString(packedBytes)
	}
	if err != nil {
		if call.Function == "()" {
			logger.InfoMsg("Calling the fallback function")
		} else {
			err = util.ABIErrorHandler(err, call, nil, logger)
			return
		}
	}

	if funcSpec.Constant {
		logger.InfoMsg("Function call to constant function, query-contract type job will be faster than call")
	}

	logger.TraceMsg("Calling",
		"destination", call.Destination,
		"function", call.Function,
		"data", callData)

	return client.Call(&def.CallArg{
		Input:    call.Source,
		Amount:   call.Amount,
		Address:  call.Destination,
		Fee:      call.Fee,
		Gas:      call.Gas,
		Data:     callData,
		Sequence: call.Sequence,
	}, logger)
}

func CallJob(call *def.Call, tx *payload.CallTx, do *def.DeployArgs, playbook *def.Playbook, client *def.Client, logger *logging.Logger) (string, []*abi.Variable, error) {
	var err error

	// Sign, broadcast, display
	txe, err := client.SignAndBroadcast(tx, logger)
	if err != nil {
		var err = util.ChainErrorHandler(payload.InputsString(tx.GetInputs()), err, logger)
		return "", nil, err
	}

	if txe.Exception != nil {
		switch txe.Exception.ErrorCode() {
		case errors.ErrorCodeExecutionReverted:
			message, err := abi.UnpackRevert(txe.Result.Return)
			if err != nil {
				return "", nil, err
			}
			if message != nil {
				logger.InfoMsg("Transaction reverted with reason",
					"Revert Reason", *message)
				return *message, nil, txe.Exception.AsError()
			} else {
				logger.InfoMsg("Transaction reverted with no reason")
				return "", nil, txe.Exception.AsError()
			}
		default:
			logger.InfoMsg("Transaction execution exception")
			return "", nil, txe.Exception.AsError()
		}
	}

	logEvents(txe, client, logger)

	var result string

	// Formally process the return
	if txe.GetResult().GetReturn() != nil {
		logger.TraceMsg("Decoding Raw Result", "return", hex.EncodeUpperToString(txe.Result.Return))

		if call.Bin != "" {
			call.Variables, err = abi.DecodeFunctionReturnFromFile(call.Bin, playbook.BinPath, call.Function, txe.Result.Return, logger)
		}
		if call.Bin == "" || err != nil {
			call.Variables, err = abi.DecodeFunctionReturnFromFile(call.Destination, playbook.BinPath, call.Function, txe.Result.Return, logger)
		}
		if err != nil {
			return "", nil, err
		}
		logger.TraceMsg("Variables", "call", call.Variables)
		result = util.GetReturnValue(call.Variables, logger)
		if result != "" {
			logger.InfoMsg("Return value", "value", result)
		} else {
			logger.InfoMsg("No return value")
		}
	} else {
		logger.InfoMsg("No return result value")
	}

	if call.Save == "tx" {
		logger.InfoMsg("Saving tx hash instead of contract return")
		result = fmt.Sprintf("%X", txe.Receipt.TxHash)
	}

	return result, call.Variables, nil
}

func deployFinalize(do *def.DeployArgs, client *def.Client, tx payload.Payload, logger *logging.Logger) (*crypto.Address, error) {
	txe, err := client.SignAndBroadcast(tx, logger)
	if err != nil {
		return nil, util.ChainErrorHandler(payload.InputsString(tx.GetInputs()), err, logger)
	}

	if err := util.ReadTxSignAndBroadcast(txe, err, logger); err != nil {
		return nil, err
	}

	// The contructor can generate events
	logEvents(txe, client, logger)

	if !txe.Receipt.CreatesContract || txe.Receipt.ContractAddress == crypto.ZeroAddress {
		// Shouldn't get ZeroAddress when CreatesContract is true, but still
		return nil, fmt.Errorf("result from SignAndBroadcast does not contain address for the deployed contract")
	}
	return &txe.Receipt.ContractAddress, nil
}

func logEvents(txe *exec.TxExecution, client *def.Client, logger *logging.Logger) {
	if client.AllSpecs == nil {
		return
	}

	for _, event := range txe.Events {
		eventLog := event.GetLog()

		if eventLog == nil {
			continue
		}

		var eventID abi.EventID
		copy(eventID[:], eventLog.GetTopic(0).Bytes())

		evAbi, ok := client.AllSpecs.EventsById[eventID]
		if !ok {
			logger.InfoMsg("Could not find ABI for Event", "Event ID", hex.EncodeUpperToString(eventID[:]))
			continue
		}

		vals := make([]interface{}, len(evAbi.Inputs))
		for i := range vals {
			vals[i] = new(string)
		}

		if err := abi.UnpackEvent(&evAbi, eventLog.Topics, eventLog.Data, vals...); err == nil {
			var fields []interface{}
			fields = append(fields, "name")
			fields = append(fields, evAbi.Name)
			for i := range vals {
				fields = append(fields, evAbi.Inputs[i].Name)
				val := vals[i].(*string)
				fields = append(fields, *val)
			}
			logger.TraceMsg("EVM Event", fields...)
		}
	}
}

func mergeAbiSpecBytes(client *def.Client, bs []byte) {
	spec, err := abi.ReadAbiSpec(bs)
	if err == nil {
		client.AllSpecs = abi.MergeAbiSpec([]*abi.AbiSpec{client.AllSpecs, spec})
	}
}
