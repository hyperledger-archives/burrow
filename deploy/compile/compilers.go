package compile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/logging"
)

type SolidityInput struct {
	Language string                         `json:"language"`
	Sources  map[string]SolidityInputSource `json:"sources"`
	Settings struct {
		Libraries map[string]map[string]string `json:"libraries"`
		Optimizer struct {
			Enabled bool `json:"enabled"`
		} `json:"optimizer"`
		OutputSelection struct {
			File struct {
				OutputType []string `json:"*"`
			} `json:"*"`
		} `json:"outputSelection"`
	} `json:"settings"`
}

type SolidityInputSource struct {
	Content string   `json:"content,omitempty"`
	Urls    []string `json:"urls,omitempty"`
}

type SolidityOutput struct {
	Contracts map[string]map[string]SolidityContract
	Errors    []struct {
		Component        string
		FormattedMessage string
		Message          string
		Severity         string
		Type             string
	}
}

type SolidityContract struct {
	Abi json.RawMessage
	Evm struct {
		Bytecode struct {
			Object         string
			Opcodes        string
			LinkReferences json.RawMessage
		}
	}
	Devdoc   json.RawMessage
	Userdoc  json.RawMessage
	Metadata string
}

type Response struct {
	Objects []ResponseItem `json:"objects"`
	Warning string         `json:"warning"`
	Version string         `json:"version"`
	Error   string         `json:"error"`
}

// Compile response object
type ResponseItem struct {
	Filename   string           `json:"filename"`
	Objectname string           `json:"objectname"`
	Contract   SolidityContract `json:"binary"`
}

func LoadSolidityContract(file string) (*SolidityContract, error) {
	codeB, err := ioutil.ReadFile(file)
	if err != nil {
		return &SolidityContract{}, err
	}
	contract := SolidityContract{}
	err = json.Unmarshal(codeB, &contract)
	if err != nil {
		return &SolidityContract{}, err
	}
	return &contract, nil
}

func (contract *SolidityContract) Save(dir, file string) error {
	str, err := json.Marshal(*contract)
	if err != nil {
		return err
	}
	// This will make the contract file appear atomically
	// This is important since if we run concurrent jobs, one job could be compiling a solidity
	// file while another reads the bin file. If write is incomplete, it will result in failures
	f, err := ioutil.TempFile(dir, "bin.*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, err = f.Write(str)
	if err != nil {
		return err
	}
	f.Close()
	return os.Rename(f.Name(), filepath.Join(dir, file))
}

func (contract *SolidityContract) Link(libraries map[string]string) error {
	bin := contract.Evm.Bytecode.Object
	if !strings.Contains(bin, "_") {
		return nil
	}
	var links map[string]map[string][]struct{ Start, Length int }
	err := json.Unmarshal(contract.Evm.Bytecode.LinkReferences, &links)
	if err != nil {
		return err
	}
	for _, f := range links {
		for name, relos := range f {
			addr, ok := libraries[name]
			if !ok {
				return fmt.Errorf("library %s is not defined", name)
			}
			for _, relo := range relos {
				if relo.Length != crypto.AddressLength {
					return fmt.Errorf("linkReference should be %d bytes long, not %d", crypto.AddressLength, relo.Length)
				}
				if len(addr) != crypto.AddressHexLength {
					return fmt.Errorf("address %s should be %d character long, not %d", addr, crypto.AddressHexLength, len(addr))
				}
				start := relo.Start * 2
				end := relo.Start*2 + crypto.AddressHexLength
				if bin[start+1] != '_' || bin[end-1] != '_' {
					return fmt.Errorf("relocation dummy not found at %d in %s ", relo.Start, bin)
				}
				bin = bin[:start] + addr + bin[end:]
			}
		}
	}

	contract.Evm.Bytecode.Object = bin

	return nil
}

func Compile(file string, optimize bool, workDir string, libraries map[string]string, logger *logging.Logger) (*Response, error) {
	input := SolidityInput{Language: "Solidity", Sources: make(map[string]SolidityInputSource)}

	input.Sources[file] = SolidityInputSource{Urls: []string{file}}
	input.Settings.Optimizer.Enabled = optimize
	input.Settings.OutputSelection.File.OutputType = []string{"abi", "evm.bytecode.linkReferences", "metadata", "bin", "devdoc"}
	input.Settings.Libraries = make(map[string]map[string]string)
	input.Settings.Libraries[""] = make(map[string]string)

	for l, a := range libraries {
		input.Settings.Libraries[""][l] = "0x" + a
	}

	command, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	logger.TraceMsg("Command Input", "command", string(command))
	result, err := runSolidity(string(command), workDir)
	if err != nil {
		return nil, err
	}
	logger.TraceMsg("Command Output", "result", result)

	output := SolidityOutput{}
	err = json.Unmarshal([]byte(result), &output)
	if err != nil {
		return nil, err
	}

	respItemArray := make([]ResponseItem, 0)

	for f, s := range output.Contracts {
		for contract, item := range s {
			respItem := ResponseItem{
				Filename:   f,
				Objectname: objectName(contract),
				Contract:   item,
			}
			respItemArray = append(respItemArray, respItem)
		}
	}

	warnings := ""
	errors := ""
	for _, msg := range output.Errors {
		if msg.Type == "Warning" {
			warnings += msg.FormattedMessage
		} else {
			errors += msg.FormattedMessage
		}
	}

	for _, re := range respItemArray {
		logger.TraceMsg("Response formulated",
			"name", re.Objectname,
			"bin", re.Contract.Evm.Bytecode.Object,
			"abi", string(re.Contract.Abi))
	}

	resp := Response{
		Objects: respItemArray,
		Warning: warnings,
		Error:   errors,
	}

	return &resp, nil
}

func objectName(contract string) string {
	if contract == "" {
		return ""
	}
	parts := strings.Split(strings.TrimSpace(contract), ":")
	return parts[len(parts)-1]
}

func runSolidity(jsonCmd string, workDir string) (string, error) {
	buf := bytes.NewBufferString(jsonCmd)
	shellCmd := exec.Command("solc", "--standard-json", "--allow-paths", "/")
	if workDir != "" {
		shellCmd.Dir = workDir
	}
	shellCmd.Stdin = buf
	output, err := shellCmd.CombinedOutput()
	s := string(output)
	return s, err
}

func PrintResponse(resp Response, cli bool, logger *logging.Logger) {
	if resp.Error != "" {
		logger.InfoMsg("solidity error", "errors", resp.Error)
	} else {
		for _, r := range resp.Objects {
			logger.InfoMsg("Response",
				"name", r.Objectname,
				"bin", r.Contract.Evm.Bytecode,
				"abi", string(r.Contract.Abi[:]),
				"link", string(r.Contract.Evm.Bytecode.LinkReferences[:]),
			)
		}
	}
}
