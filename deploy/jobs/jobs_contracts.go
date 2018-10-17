package jobs

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hyperledger/burrow/crypto"
	compilers "github.com/hyperledger/burrow/deploy/compile"
	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/util"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/txs/payload"
	log "github.com/sirupsen/logrus"
)

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
		b, err := json.Marshal(res.Binary)
		if err != nil {
			return "", err
		}
		contractName := filepath.Join(binPath, fmt.Sprintf("%s.bin", res.Objectname))
		log.WithField("=>", contractName).Warn("Saving Binary")
		if err := ioutil.WriteFile(contractName, b, 0664); err != nil {
			return "", err
		}
	}

	return "", nil
}

func DeployJob(deploy *def.Deploy, do *def.DeployArgs, client *def.Client, resp *compilers.Response) (result string, err error) {
	deploy.Libraries, _ = util.PreProcessLibs(deploy.Libraries, do, client)
	// trim the extension
	contractName := strings.TrimSuffix(deploy.Contract, filepath.Ext(deploy.Contract))

	// Use defaults
	deploy.Source = useDefault(deploy.Source, do.Package.Account)
	deploy.Instance = useDefault(deploy.Instance, contractName)
	deploy.Amount = useDefault(deploy.Amount, do.DefaultAmount)
	deploy.Fee = useDefault(deploy.Fee, do.DefaultFee)
	deploy.Gas = useDefault(deploy.Gas, do.DefaultGas)

	// assemble contract
	contractPath, err := findContractFile(deploy.Contract, do.BinPath)
	if err != nil {
		return
	}

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
				return "", fmt.Errorf("library %s should be contract:address format", l)
			}
			libs[v[0]] = v[1]
		}
	}

	// compile
	if filepath.Ext(deploy.Contract) != ".sol" {
		log.Info("Binary file detected. Using binary deploy sequence.")
		log.WithField("=>", contractPath).Info("Binary path")

		binaryResponse, err := compilers.LinkFile(contractPath, libs)
		if err != nil {
			return "", fmt.Errorf("Something went wrong with your binary deployment: %v", err)
		}
		if binaryResponse.Error != "" {
			return "", fmt.Errorf("Something went wrong when you were trying to link your binaries: %v", binaryResponse.Error)
		}
		if binaryResponse.Binary == "" {
			return "", fmt.Errorf("Contract binary code missing")
		}
		contractCode := binaryResponse.Binary

		if deploy.Data != nil {
			_, callDataArray, err := util.PreProcessInputData("", deploy.Data, do, client, true)
			if err != nil {
				return "", err
			}
			packedBytes, err := abi.ReadAbiFormulateCall(binaryResponse.Abi, "", callDataArray)
			if err != nil {
				return "", err
			}
			callData := hex.EncodeToString(packedBytes)
			contractCode = contractCode + callData
		}

		tx, err := deployTx(client, deploy, contractName, string(contractCode))
		if err != nil {
			return "could not deploy binary contract", err
		}
		result, err := deployFinalize(do, client, tx)
		if err != nil {
			return "", fmt.Errorf("Error finalizing contract deploy from path %s: %v", contractPath, err)
		}
		return result.String(), err
	} else {
		contractPath = deploy.Contract
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
		// loop through objects returned from compiler
		switch {
		case len(resp.Objects) == 1:
			log.WithField("path", contractPath).Info("Deploying the only contract in file")
			response := resp.Objects[0]
			log.WithField("=>", string(response.Binary.Abi)).Info("Abi")
			log.WithField("=>", response.Binary.Evm.Bytecode.Object).Info("Bin")
			if response.Binary.Evm.Bytecode.Object == "" {
				return "", fmt.Errorf("Contract binary code missing")
			}

			result, err = deployContract(deploy, do, client, response, libs)
			if err != nil {
				return "", err
			}
		case deploy.Instance == "all":
			log.WithField("path", contractPath).Info("Deploying all contracts")
			var baseObj string
			deployedCount := 0
			for _, response := range resp.Objects {
				if response.Binary.Evm.Bytecode.Object == "" {
					continue
				}
				result, err = deployContract(deploy, do, client, response, libs)
				if err != nil {
					return "", err
				}
				deployedCount++
				if strings.ToLower(response.Objectname) == strings.ToLower(strings.TrimSuffix(filepath.Base(deploy.Contract), filepath.Ext(filepath.Base(deploy.Contract)))) {
					baseObj = result
				}
			}
			if deployedCount == 0 {
				return "", fmt.Errorf("Contract binary code missing")
			}
			if baseObj != "" {
				result = baseObj
			}
		default:
			log.WithField("contract", deploy.Instance).Info("Deploying a single contract")
			for _, response := range resp.Objects {
				if response.Binary.Evm.Bytecode.Object == "" ||
					response.Filename != deploy.Contract {
					continue
				}
				if matchInstanceName(response.Objectname, deploy.Instance) {
					if response.Binary.Evm.Bytecode.Object == "" {
						return "", fmt.Errorf("Contract binary code missing")
					}
					log.WithField("=>", string(response.Binary.Abi)).Info("Abi")
					log.WithField("=>", response.Binary.Evm.Bytecode.Object).Info("Bin")
					result, err = deployContract(deploy, do, client, response, libs)
					if err != nil {
						return "", err
					}
				}
			}
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
func deployContract(deploy *def.Deploy, do *def.DeployArgs, client *def.Client, compilersResponse compilers.ResponseItem, libs map[string]string) (string, error) {
	log.WithField("=>", string(compilersResponse.Binary.Abi)).Debug("Specification (From Compilers)")

	linked, err := compilers.LinkContract(compilersResponse.Binary, libs)
	if err != nil {
		return "", err
	}
	contractCode := linked.Binary

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

	if deploy.Data != nil {
		_, callDataArray, err := util.PreProcessInputData(compilersResponse.Objectname, deploy.Data, do, client, true)
		if err != nil {
			return "", err
		}
		packedBytes, err := abi.ReadAbiFormulateCall(compilersResponse.Binary.Abi, "", callDataArray)
		if err != nil {
			return "", err
		}
		callData := hex.EncodeToString(packedBytes)
		contractCode = contractCode + callData
	}

	tx, err := deployTx(client, deploy, compilersResponse.Objectname, contractCode)
	if err != nil {
		return "", err
	}

	// Sign, broadcast, display
	contractAddress, err := deployFinalize(do, client, tx)
	if err != nil {
		return "", fmt.Errorf("Error finalizing contract deploy %s: %v", deploy.Contract, err)
	}

	// saving contract/library abi at abi/address
	if contractAddress != nil {
		// saving binary
		b, err := json.Marshal(compilersResponse.Binary)
		if err != nil {
			return "", err
		}
		addressBin := filepath.Join(do.BinPath, contractAddress.String())
		log.WithField("=>", addressBin).Debug("Saving Binary")
		if err := ioutil.WriteFile(addressBin, b, 0664); err != nil {
			return "", err
		}
		contractName := filepath.Join(do.BinPath, fmt.Sprintf("%s.bin", compilersResponse.Objectname))
		log.WithField("=>", contractName).Warn("Saving Binary")
		if err := ioutil.WriteFile(contractName, b, 0664); err != nil {
			return "", err
		}
		return contractAddress.String(), nil
	} else {
		// we shouldn't reach this point because we should have an error before this.
		return "", fmt.Errorf("The contract did not deploy. Unable to save abi to abi/contractAddress.")
	}
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

func CallJob(call *def.Call, do *def.DeployArgs, client *def.Client) (string, []*abi.Variable, error) {
	var err error
	var callData string
	var callDataArray []string
	//todo: find a way to call the fallback function here
	call.Function, callDataArray, err = util.PreProcessInputData(call.Function, call.Data, do, client, false)
	if err != nil {
		return "", nil, err
	}
	// Use default
	call.Source = useDefault(call.Source, do.Package.Account)
	call.Amount = useDefault(call.Amount, do.DefaultAmount)
	call.Fee = useDefault(call.Fee, do.DefaultFee)
	call.Gas = useDefault(call.Gas, do.DefaultGas)

	// formulate call
	var packedBytes []byte
	if call.Bin != "" {
		packedBytes, err = abi.ReadAbiFormulateCallFile(call.Bin, do.BinPath, call.Function, callDataArray)
		callData = hex.EncodeToString(packedBytes)
	}
	if call.Bin == "" || err != nil {
		packedBytes, err = abi.ReadAbiFormulateCallFile(call.Destination, do.BinPath, call.Function, callDataArray)
		callData = hex.EncodeToString(packedBytes)
	}
	if err != nil {
		if call.Function == "()" {
			log.Warn("Calling the fallback function")
		} else {
			var str, err = util.ABIErrorHandler(do, err, call, nil)
			return str, nil, err
		}
	}

	log.WithFields(log.Fields{
		"destination": call.Destination,
		"function":    call.Function,
		"data":        callData,
	}).Info("Calling")

	tx, err := client.Call(&def.CallArg{
		Input:    call.Source,
		Amount:   call.Amount,
		Address:  call.Destination,
		Fee:      call.Fee,
		Gas:      call.Gas,
		Data:     callData,
		Sequence: call.Sequence,
	})
	if err != nil {
		return "", nil, err
	}

	// Sign, broadcast, display
	txe, err := client.SignAndBroadcast(tx)
	if err != nil {
		var err = util.ChainErrorHandler(do.Package.Account, err)
		return "", nil, err
	}

	if txe.Exception != nil && txe.Exception.ErrorCode() == errors.ErrorCodeExecutionReverted {
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
	}
	var result string
	log.Debug(txe.Result.Return)

	// Formally process the return
	if txe.Result.Return != nil {
		log.WithField("=>", result).Debug("Decoding Raw Result")
		if call.Bin != "" {
			call.Variables, err = abi.ReadAndDecodeContractReturn(call.Bin, do.BinPath, call.Function, txe.Result.Return)
		}
		if call.Bin == "" || err != nil {
			call.Variables, err = abi.ReadAndDecodeContractReturn(call.Destination, do.BinPath, call.Function, txe.Result.Return)
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

func deployFinalize(do *def.DeployArgs, client *def.Client, tx payload.Payload) (*crypto.Address, error) {
	txe, err := client.SignAndBroadcast(tx)
	if err != nil {
		return nil, util.ChainErrorHandler(do.Package.Account, err)
	}

	if err := util.ReadTxSignAndBroadcast(txe, err); err != nil {
		return nil, err
	}

	if !txe.Receipt.CreatesContract || txe.Receipt.ContractAddress == crypto.ZeroAddress {
		// Shouldn't get ZeroAddress when CreatesContract is true, but still
		return nil, fmt.Errorf("result from SignAndBroadcast does not contain address for the deployed contract")
	}
	return &txe.Receipt.ContractAddress, nil
}
