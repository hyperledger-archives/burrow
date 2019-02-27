package jobs

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"

	"github.com/hyperledger/burrow/crypto"
	compilers "github.com/hyperledger/burrow/deploy/compile"
	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/util"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/txs/payload"
	log "github.com/sirupsen/logrus"
)

var errCodeMissing = fmt.Errorf("error: no binary code found in contract. Contract may be abstract due to missing function body or inherited function signatures not matching.")

func BuildJob(build *def.Build, binPath string, resp *compilers.Response) (result string, err error) {
	// assemble contract
	contractPath, err := findContractFile(build.Contract, binPath)
	if err != nil {
		return
	}

	log.WithField("=>", contractPath).Info("Contract path")

	// normal compilation/deploy sequence
	if resp == nil {
		log.Errorln("Error compiling contracts: Missing compiler result")
		return "", fmt.Errorf("internal error")
	} else if resp.Error != "" {
		log.Errorln("Error compiling contracts: Language error:")
		return "", fmt.Errorf("%v", resp.Error)
	} else if resp.Warning != "" {
		log.WithField("=>", resp.Warning).Warn("Warning during contract compilation")
	}

	// Save
	binP := build.BinPath
	if binP == "" {
		binP = binPath
	}
	if _, err := os.Stat(binP); os.IsNotExist(err) {
		if err := os.Mkdir(binP, 0775); err != nil {
			return "", err
		}
	}

	for _, res := range resp.Objects {
		switch build.Instance {
		case "":
			if res.Filename != contractPath {
				continue
			}
		case "all":
		default:
			if res.Objectname != build.Instance {
				continue
			}
		}

		// saving binary
		b, err := json.Marshal(res.Contract)
		if err != nil {
			return "", err
		}
		contractName := filepath.Join(binP, fmt.Sprintf("%s.bin", res.Objectname))
		log.WithField("=>", contractName).Warn("Saving Binary")
		if err := ioutil.WriteFile(contractName, b, 0664); err != nil {
			return "", err
		}
	}

	return "", nil
}

func FormulateDeployJob(deploy *def.Deploy, do *def.DeployArgs, deployScript *def.Playbook, client *def.Client, intermediate interface{}) (txs []*payload.CallTx, contracts []*compilers.ResponseItem, err error) {
	deploy.Libraries, _ = util.PreProcessLibs(deploy.Libraries, do, deployScript, client)
	// trim the extension and path
	contractName := filepath.Base(deploy.Contract)
	contractName = strings.TrimSuffix(contractName, filepath.Ext(contractName))

	// Use defaults
	deploy.Source = useDefault(deploy.Source, deployScript.Account)
	deploy.Instance = useDefault(deploy.Instance, contractName)
	deploy.Amount = useDefault(deploy.Amount, do.DefaultAmount)
	deploy.Fee = useDefault(deploy.Fee, do.DefaultFee)
	deploy.Gas = useDefault(deploy.Gas, do.DefaultGas)

	// assemble contract
	contractPath, err := findContractFile(deploy.Contract, do.BinPath)
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
		log.Info("Binary file detected. Using binary deploy sequence.")
		log.WithField("=>", contractPath).Info("Binary path")

		contract, err := compilers.LoadSolidityContract(contractPath)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to read contract %s: %v", contractPath, err)
		}
		err = contract.Link(libs)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to link contract %s: %v", contractPath, err)
		}
		contractCode := contract.Evm.Bytecode.Object

		mergeAbiSpecBytes(do, contract.Abi)

		if deploy.Data != nil {
			_, callDataArray, err := util.PreProcessInputData("", deploy.Data, do, deployScript, client, true)
			if err != nil {
				return nil, nil, err
			}
			packedBytes, _, err := abi.EncodeFunctionCall(string(contract.Abi), "", callDataArray...)
			if err != nil {
				return nil, nil, err
			}
			callData := hex.EncodeToString(packedBytes)
			contractCode = contractCode + callData
		}

		tx, err := deployTx(client, deploy, contractName, string(contractCode))
		if err != nil {
			return nil, nil, fmt.Errorf("could not deploy binary contract: %v", err)
		}
		txs = []*payload.CallTx{tx}
		contracts = append(contracts, &compilers.ResponseItem{Filename: contractPath, Objectname: contractName, Contract: *contract})
	} else {
		contractPath = deploy.Contract
		log.WithField("=>", contractPath).Info("Contract path")
		// normal compilation/deploy sequence

		resp, err := getCompilerWork(intermediate)
		if err != nil {
			return nil, nil, err
		}

		if resp == nil {
			log.Errorln("Error compiling contracts: Missing compiler result")
			return nil, nil, fmt.Errorf("internal error")
		} else if resp.Error != "" {
			log.Errorln("Error compiling contracts: Language error:")
			return nil, nil, fmt.Errorf("%v", resp.Error)
		} else if resp.Warning != "" {
			log.WithField("=>", resp.Warning).Warn("Warning during contract compilation")
		}
		// loop through objects returned from compiler
		switch {
		case len(resp.Objects) == 1:
			log.WithField("path", contractPath).Info("Deploying the only contract in file")
			response := resp.Objects[0]
			log.WithField("=>", string(response.Contract.Abi)).Info("Abi")
			log.WithField("=>", response.Contract.Evm.Bytecode.Object).Info("Bin")
			if response.Contract.Evm.Bytecode.Object == "" {
				return nil, nil, errCodeMissing
			}
			mergeAbiSpecBytes(do, response.Contract.Abi)

			tx, err := deployContract(deploy, do, deployScript, client, response, libs)
			if err != nil {
				return nil, nil, err
			}

			txs = []*payload.CallTx{tx}
			contracts = append(contracts, &resp.Objects[0])
		case deploy.Instance == "all":
			log.WithField("path", contractPath).Info("Deploying all contracts")
			var baseObj *payload.CallTx
			var baseContract *compilers.ResponseItem
			deployedCount := 0
			for i, response := range resp.Objects {
				if response.Contract.Evm.Bytecode.Object == "" {
					continue
				}
				mergeAbiSpecBytes(do, response.Contract.Abi)
				tx, err := deployContract(deploy, do, deployScript, client, response, libs)
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
			log.WithField("contract", deploy.Instance).Info("Deploying a single contract")
			for i, response := range resp.Objects {
				if response.Contract.Evm.Bytecode.Object == "" ||
					response.Filename != deploy.Contract {
					continue
				}
				if matchInstanceName(response.Objectname, deploy.Instance) {
					if response.Contract.Evm.Bytecode.Object == "" {
						return nil, nil, errCodeMissing
					}
					log.WithField("contract", response.Objectname).Infof("foo %s", deploy.Instance)
					log.WithField("=>", string(response.Contract.Abi)).Info("Abi")
					log.WithField("=>", response.Contract.Evm.Bytecode.Object).Info("Bin")
					tx, err := deployContract(deploy, do, deployScript, client, response, libs)
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

func DeployJob(deploy *def.Deploy, do *def.DeployArgs, script *def.Playbook, client *def.Client, txs []*payload.CallTx, contracts []*compilers.ResponseItem) (result string, err error) {

	// Save
	if _, err := os.Stat(do.BinPath); os.IsNotExist(err) {
		if err := os.Mkdir(do.BinPath, 0775); err != nil {
			return "", err
		}
	}

	// saving contract
	// additional data may be sent along with the contract
	// these are naively added to the end of the contract code using standard
	// mint packing

	for i, tx := range txs {
		// Sign, broadcast, display
		contractAddress, err := deployFinalize(do, script, client, tx)
		if err != nil {
			return "", fmt.Errorf("Error finalizing contract deploy %s: %v", deploy.Contract, err)
		}

		// saving contract/library abi at abi/address
		if contracts != nil && contractAddress != nil {
			contract := contracts[i].Contract
			// saving binary
			addressBin := filepath.Join(do.BinPath, contractAddress.String())
			log.WithField("=>", addressBin).Debug("Saving Binary")
			err = contract.Save(addressBin)
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

func findContractFile(contract, binPath string) (string, error) {
	contractPaths := []string{contract, filepath.Join(binPath, contract), filepath.Join(binPath, filepath.Base(contract))}

	for _, p := range contractPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("Could not find contract in any of %v", contractPaths)
}

// TODO [rj] refactor to remove [contractPath] from functions signature => only used in a single error throw.
func deployContract(deploy *def.Deploy, do *def.DeployArgs, script *def.Playbook, client *def.Client, compilersResponse compilers.ResponseItem, libs map[string]string) (*payload.CallTx, error) {
	log.WithField("=>", string(compilersResponse.Contract.Abi)).Debug("Specification (From Compilers)")

	contract := compilersResponse.Contract
	contractName := filepath.Join(do.BinPath, fmt.Sprintf("%s.bin", compilersResponse.Objectname))
	log.WithField("=>", contractName).Warn("Saving Binary")
	err := contract.Save(contractName)
	if err != nil {
		return nil, err
	}

	err = contract.Link(libs)
	if err != nil {
		return nil, err
	}
	contractCode := contract.Evm.Bytecode.Object

	if deploy.Data != nil {
		_, callDataArray, err := util.PreProcessInputData(compilersResponse.Objectname, deploy.Data, do, script, client, true)
		if err != nil {
			return nil, err
		}
		packedBytes, _, err := abi.EncodeFunctionCall(string(compilersResponse.Contract.Abi), "", callDataArray...)
		if err != nil {
			return nil, err
		}
		callData := hex.EncodeToString(packedBytes)
		contractCode = contractCode + callData
	}

	return deployTx(client, deploy, compilersResponse.Objectname, contractCode)
}

func deployTx(client *def.Client, deploy *def.Deploy, contractName, contractCode string) (*payload.CallTx, error) {
	// Deploy contract
	log.WithFields(log.Fields{
		"name": contractName,
	}).Warn("Deploying Contract")

	log.WithFields(log.Fields{
		"source":    deploy.Source,
		"code":      contractCode,
		"chain-url": client.ChainAddress,
	}).Info()

	return client.Call(&def.CallArg{
		Input:    deploy.Source,
		Amount:   deploy.Amount,
		Fee:      deploy.Fee,
		Gas:      deploy.Gas,
		Data:     contractCode,
		Sequence: deploy.Sequence,
	})
}

func FormulateCallJob(call *def.Call, do *def.DeployArgs, deployScript *def.Playbook, client *def.Client) (tx *payload.CallTx, err error) {
	var callData string
	var callDataArray []string
	//todo: find a way to call the fallback function here
	call.Function, callDataArray, err = util.PreProcessInputData(call.Function, call.Data, do, deployScript, client, false)
	if err != nil {
		return nil, err
	}
	// Use default
	call.Source = useDefault(call.Source, deployScript.Account)
	call.Amount = useDefault(call.Amount, do.DefaultAmount)
	call.Fee = useDefault(call.Fee, do.DefaultFee)
	call.Gas = useDefault(call.Gas, do.DefaultGas)

	// formulate call
	var packedBytes []byte
	var constant bool
	if call.Bin != "" {
		packedBytes, constant, err = abi.EncodeFunctionCallFromFile(call.Bin, do.BinPath, call.Function, callDataArray)
		callData = hex.EncodeToString(packedBytes)
	}
	if call.Bin == "" || err != nil {
		packedBytes, constant, err = abi.EncodeFunctionCallFromFile(call.Destination, do.BinPath, call.Function, callDataArray)
		callData = hex.EncodeToString(packedBytes)
	}
	if err != nil {
		if call.Function == "()" {
			log.Warn("Calling the fallback function")
		} else {
			err = util.ABIErrorHandler(err, call, nil)
			return
		}
	}

	if constant {
		log.Warn("Function call to constant function, query-contract type job will be faster than call")
	}

	log.WithFields(log.Fields{
		"destination": call.Destination,
		"function":    call.Function,
		"data":        callData,
	}).Info("Calling")

	return client.Call(&def.CallArg{
		Input:    call.Source,
		Amount:   call.Amount,
		Address:  call.Destination,
		Fee:      call.Fee,
		Gas:      call.Gas,
		Data:     callData,
		Sequence: call.Sequence,
	})
}

func CallJob(call *def.Call, tx *payload.CallTx, do *def.DeployArgs, deployScript *def.Playbook, client *def.Client) (string, []*abi.Variable, error) {
	var err error

	// Sign, broadcast, display
	txe, err := client.SignAndBroadcast(tx)
	if err != nil {
		var err = util.ChainErrorHandler(deployScript.Account, err)
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
				log.WithField("Revert Reason", *message).Error("Transaction reverted with reason")
				return *message, nil, txe.Exception.AsError()
			} else {
				log.Error("Transaction reverted with no reason")
				return "", nil, txe.Exception.AsError()
			}
		default:
			log.Error("Transaction execution exception")
			return "", nil, txe.Exception.AsError()
		}
	}

	logEvents(txe, do)

	var result string

	// Formally process the return
	if txe.GetResult().GetReturn() != nil {
		log.Debug(txe.Result.Return)

		log.WithField("=>", result).Debug("Decoding Raw Result")
		if call.Bin != "" {
			call.Variables, err = abi.DecodeFunctionReturnFromFile(call.Bin, do.BinPath, call.Function, txe.Result.Return)
		}
		if call.Bin == "" || err != nil {
			call.Variables, err = abi.DecodeFunctionReturnFromFile(call.Destination, do.BinPath, call.Function, txe.Result.Return)
		}
		if err != nil {
			return "", nil, err
		}
		log.WithField("=>", call.Variables).Debug("call variables:")
		result = util.GetReturnValue(call.Variables)
		if result != "" {
			log.WithField("=>", result).Warn("Return Value")
		} else {
			log.Debug("No return.")
		}
	} else {
		log.Debug("No return from contract.")
	}

	if call.Save == "tx" {
		log.Info("Saving tx hash instead of contract return")
		result = fmt.Sprintf("%X", txe.Receipt.TxHash)
	}

	return result, call.Variables, nil
}

func deployFinalize(do *def.DeployArgs, deployScript *def.Playbook, client *def.Client, tx payload.Payload) (*crypto.Address, error) {
	txe, err := client.SignAndBroadcast(tx)
	if err != nil {
		return nil, util.ChainErrorHandler(deployScript.Account, err)
	}

	if err := util.ReadTxSignAndBroadcast(txe, err); err != nil {
		return nil, err
	}

	// The contructor can generate events
	logEvents(txe, do)

	if !txe.Receipt.CreatesContract || txe.Receipt.ContractAddress == crypto.ZeroAddress {
		// Shouldn't get ZeroAddress when CreatesContract is true, but still
		return nil, fmt.Errorf("result from SignAndBroadcast does not contain address for the deployed contract")
	}
	return &txe.Receipt.ContractAddress, nil
}

func logEvents(txe *exec.TxExecution, do *def.DeployArgs) {
	if do.AllSpecs == nil {
		return
	}

	for _, event := range txe.Events {
		eventLog := event.GetLog()

		if eventLog == nil {
			continue
		}

		var eventID abi.EventID
		copy(eventID[:], eventLog.GetTopic(0).Bytes())

		evAbi, ok := do.AllSpecs.EventsById[eventID]
		if !ok {
			log.Errorf("Could not find ABI for Event with ID %x\n", eventID)
			continue
		}

		vals := make([]interface{}, len(evAbi.Inputs))
		for i := range vals {
			vals[i] = new(string)
		}

		if err := abi.UnpackEvent(&evAbi, eventLog.Topics, eventLog.Data, vals...); err == nil {
			fields := log.Fields{}
			for i := range vals {
				val := vals[i].(*string)
				fields[evAbi.Inputs[i].Name] = *val
			}
			log.WithFields(fields).Info("Event " + evAbi.Name)
		}
	}
}

func mergeAbiSpecBytes(do *def.DeployArgs, bs []byte) {
	spec, err := abi.ReadAbiSpec(bs)
	if err == nil {
		do.AllSpecs = abi.MergeAbiSpec([]*abi.AbiSpec{do.AllSpecs, spec})
	}
}
