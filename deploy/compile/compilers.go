package compile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"

	"github.com/hyperledger/burrow/crypto"
	log "github.com/sirupsen/logrus"
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
	Contracts map[string]map[string]SolidityOutputContract
	Errors    []struct {
		Component        string
		FormattedMessage string
		Message          string
		Severity         string
		Type             string
	}
}

type SolidityOutputContract struct {
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

type BinaryResponse struct {
	Binary string          `json:"binary"`
	Abi    json.RawMessage `json:"abi"`
	Error  string          `json:"error"`
}

// Compile response object
type ResponseItem struct {
	Filename   string                 `json:"filename"`
	Objectname string                 `json:"objectname"`
	Binary     SolidityOutputContract `json:"binary"`
}

func LinkFile(file string, libraries map[string]string) (*BinaryResponse, error) {
	//Create Binary Request, send it off
	codeB, err := ioutil.ReadFile(file)
	if err != nil {
		return &BinaryResponse{}, err
	}
	contract := SolidityOutputContract{}
	err = json.Unmarshal(codeB, &contract)
	if err != nil {
		return &BinaryResponse{}, err
	}
	return LinkContract(contract, libraries)
}

func LinkContract(contract SolidityOutputContract, libraries map[string]string) (*BinaryResponse, error) {
	bin := contract.Evm.Bytecode.Object
	if !strings.Contains(bin, "_") {
		return &BinaryResponse{
			Binary: bin,
			Abi:    contract.Abi,
			Error:  "",
		}, nil
	}
	var links map[string]map[string][]struct{ Start, Length int }
	err := json.Unmarshal(contract.Evm.Bytecode.LinkReferences, &links)
	if err != nil {
		return &BinaryResponse{}, err
	}
	for _, f := range links {
		for name, relos := range f {
			addr, ok := libraries[name]
			if !ok {
				return &BinaryResponse{}, fmt.Errorf("library %s is not defined", name)
			}
			for _, relo := range relos {
				if relo.Length != crypto.AddressLength {
					return &BinaryResponse{}, fmt.Errorf("linkReference should be %d bytes long, not %d", crypto.AddressLength, relo.Length)
				}
				if len(addr) != crypto.AddressHexLength {
					return &BinaryResponse{}, fmt.Errorf("address %s should be %d character long, not %d", addr, crypto.AddressHexLength, len(addr))
				}
				start := relo.Start * 2
				end := relo.Start*2 + crypto.AddressHexLength
				if bin[start+1] != '_' || bin[end-1] != '_' {
					return &BinaryResponse{}, fmt.Errorf("relocation dummy not found at %d in %s ", relo.Start, bin)
				}
				bin = bin[:start] + addr + bin[end:]
			}
		}
	}

	return &BinaryResponse{
		Binary: bin,
		Abi:    contract.Abi,
		Error:  "",
	}, nil
}

func Compile(file string, optimize bool, libraries map[string]string) (*Response, error) {
	input := SolidityInput{Language: "Solidity", Sources: make(map[string]SolidityInputSource)}

	input.Sources[file] = SolidityInputSource{Urls: []string{file}}
	input.Settings.Optimizer.Enabled = optimize
	input.Settings.OutputSelection.File.OutputType = []string{"abi", "evm.bytecode.linkReferences", "metadata", "bin", "devdoc"}
	input.Settings.Libraries = make(map[string]map[string]string)
	input.Settings.Libraries[""] = make(map[string]string)

	if libraries != nil {
		for l, a := range libraries {
			input.Settings.Libraries[""][l] = "0x" + a
		}
	}

	command, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	log.WithField("Command: ", string(command)).Debug("Command Input")
	result, err := runSolidity(string(command))
	if err != nil {
		return nil, err
	}
	log.WithField("Command Result: ", result).Debug("Command Output")

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
				Binary:     item,
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
		log.WithFields(log.Fields{
			"name": re.Objectname,
			"bin":  re.Binary.Evm.Bytecode.Object,
			"abi":  string(re.Binary.Abi),
		}).Debug("Response formulated")
	}

	resp := Response{
		Objects: respItemArray,
		Warning: warnings,
		Error:   errors,
	}

	PrintResponse(resp, false)

	return &resp, nil
}

func objectName(contract string) string {
	if contract == "" {
		return ""
	}
	parts := strings.Split(strings.TrimSpace(contract), ":")
	return parts[len(parts)-1]
}

func runSolidity(jsonCmd string) (string, error) {
	buf := bytes.NewBufferString(jsonCmd)
	shellCmd := exec.Command("solc", "--standard-json", "--allow-paths", "/")
	shellCmd.Stdin = buf
	output, err := shellCmd.CombinedOutput()
	s := string(output)
	return s, err
}

func PrintResponse(resp Response, cli bool) {
	if resp.Error != "" {
		log.Warn(resp.Error)
	} else {
		for _, r := range resp.Objects {
			message := log.WithFields((log.Fields{
				"name": r.Objectname,
				"bin":  r.Binary.Evm.Bytecode,
				"abi":  string(r.Binary.Abi[:]),
				"link": string(r.Binary.Evm.Bytecode.LinkReferences[:]),
			}))
			if cli {
				message.Warn("Response")
			} else {
				message.Info("Response")
			}
		}
	}
}

func extractObjectNames(script []byte) ([]string, error) {
	regExpression, err := regexp.Compile("(contract|library) (.+?) (is)?(.+?)?({)")
	if err != nil {
		return nil, err
	}
	objectNamesList := regExpression.FindAllSubmatch(script, -1)
	var objects []string
	for _, objectNames := range objectNamesList {
		objects = append(objects, string(objectNames[2]))
	}
	return objects, nil
}
