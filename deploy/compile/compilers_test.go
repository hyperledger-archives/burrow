package compile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// full solc response object
// individual contract items
type SolcItem struct {
	Bin string `json:"bin"`
	Abi string `json:"abi"`
}

type SolcResponse struct {
	Contracts map[string]*SolcItem `mapstructure:"contracts" json:"contracts"`
	Version   string               `mapstructure:"version" json:"version"` // json encoded
}

func BlankSolcItem() *SolcItem {
	return &SolcItem{}
}

func BlankSolcResponse() *SolcResponse {
	return &SolcResponse{
		Version:   "",
		Contracts: make(map[string]*SolcItem),
	}
}

func TestLocalMulti(t *testing.T) {
	os.Chdir(testContractPath()) // important to maintain relative paths

	expectedSolcResponse := BlankSolcResponse()
	actualOutput, err := exec.Command("solc", "--combined-json", "bin,abi", "contractImport1.sol").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	warning, responseJSON := extractWarningJSON(strings.TrimSpace(string(actualOutput)))
	err = json.Unmarshal([]byte(responseJSON), expectedSolcResponse)

	respItemArray := make([]ResponseItem, 0)

	for contract, item := range expectedSolcResponse.Contracts {
		respItem := ResponseItem{
			Objectname: objectName(strings.TrimSpace(contract)),
		}
		respItem.Binary.Evm.Bytecode.Object = item.Bin
		respItemArray = append(respItemArray, respItem)
	}
	expectedResponse := &Response{
		Objects: respItemArray,
		Warning: warning,
		Version: "",
		Error:   "",
	}
	resp, err := RequestCompile("contractImport1.sol", false, make(map[string]string))
	if err != nil {
		t.Fatal(err)
	}
	allClear := true
	for _, object := range expectedResponse.Objects {
		if !contains(resp.Objects, object) {
			allClear = false
		}
	}
	if !allClear {
		t.Errorf("Got incorrect response, expected %v, \n\n got %v", expectedResponse, resp)
	}
}

func TestLocalSingle(t *testing.T) {
	os.Chdir(testContractPath()) // important to maintain relative paths

	expectedSolcResponse := BlankSolcResponse()

	shellCmd := exec.Command("solc", "--combined-json", "bin,abi", "simpleContract.sol")
	actualOutput, err := shellCmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	warning, responseJSON := extractWarningJSON(strings.TrimSpace(string(actualOutput)))
	err = json.Unmarshal([]byte(responseJSON), expectedSolcResponse)

	respItemArray := make([]ResponseItem, 0)

	for contract, item := range expectedSolcResponse.Contracts {
		respItem := ResponseItem{
			Objectname: objectName(strings.TrimSpace(contract)),
			Filename:   "simpleContract.sol",
		}
		respItem.Binary.Abi = json.RawMessage(item.Abi)
		respItem.Binary.Evm.Bytecode.Object = item.Bin
		respItem.Binary.Evm.Bytecode.LinkReferences = []byte("{}")
		respItemArray = append(respItemArray, respItem)
	}
	expectedResponse := &Response{
		Objects: respItemArray,
		Warning: warning,
		Version: "",
		Error:   "",
	}
	resp, err := RequestCompile("simpleContract.sol", false, make(map[string]string))
	if err != nil {
		t.Fatal(err)
	}
	for i := range resp.Objects {
		resp.Objects[i].Binary.Metadata = ""
		resp.Objects[i].Binary.Devdoc = nil
		resp.Objects[i].Binary.Evm.Bytecode.Opcodes = ""
	}
	assert.Equal(t, expectedResponse, resp)
}

func TestFaultyContract(t *testing.T) {
	var expectedSolcResponse Response

	actualOutput, err := exec.Command("solc", "--combined-json", "bin,abi", "faultyContract.sol").CombinedOutput()
	err = json.Unmarshal(actualOutput, expectedSolcResponse)
	t.Log(expectedSolcResponse.Error)
	resp, err := RequestCompile("faultyContract.sol", false, make(map[string]string))
	t.Log(resp.Error)
	if err != nil {
		if expectedSolcResponse.Error != resp.Error {
			t.Errorf("Expected %v got %v", expectedSolcResponse.Error, resp.Error)
		}
	}
	output := strings.TrimSpace(string(actualOutput))
	err = json.Unmarshal([]byte(output), expectedSolcResponse)
}

func testContractPath() string {
	baseDir, _ := os.Getwd()
	return filepath.Join(baseDir, "..", "..", "tests", "compilers_fixtures")
}

// The solidity 0.4.21 compiler appends something called auxdata to the end of the bin file (this is visible with
// solc --asm). This is a swarm hash of the metadata, and it's always at the end. This includes the path of the
// solidity source file, so it will differ.
func trimAuxdata(bin string) string {
	return bin[:len(bin)-86]
}

func extractWarningJSON(output string) (warning string, json string) {
	jsonBeginsCertainly := strings.Index(output, `{"contracts":`)

	if jsonBeginsCertainly > 0 {
		warning = output[:jsonBeginsCertainly]
		json = output[jsonBeginsCertainly:]
	} else {
		json = output
	}
	return
}

func contains(s []ResponseItem, e ResponseItem) bool {
	for _, a := range s {
		if a.Binary.Evm.Bytecode.Object == e.Binary.Evm.Bytecode.Object {
			return true
		}
	}
	return false
}
