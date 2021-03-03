package compile

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hyperledger/burrow/logging"
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
		t.Fatalf("solc failed %v: %s", err, actualOutput)
	}

	warning, responseJSON := extractWarningJSON(strings.TrimSpace(string(actualOutput)))
	err = json.Unmarshal([]byte(responseJSON), expectedSolcResponse)
	require.NoError(t, err)

	respItemArray := make([]ResponseItem, 0)

	for contract, item := range expectedSolcResponse.Contracts {
		respItem := ResponseItem{
			Objectname: objectName(strings.TrimSpace(contract)),
		}
		respItem.Contract.Evm.Bytecode.Object = item.Bin
		respItemArray = append(respItemArray, respItem)
	}
	expectedResponse := &Response{
		Objects: respItemArray,
		Warning: warning,
		Version: "",
		Error:   "",
	}
	resp, err := EVM("contractImport1.sol", false, "", make(map[string]string), logging.NewNoopLogger())
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
		t.Fatalf("solc failed %v: %s", err, actualOutput)
	}

	warning, responseJSON := extractWarningJSON(strings.TrimSpace(string(actualOutput)))
	err = json.Unmarshal([]byte(responseJSON), expectedSolcResponse)
	require.NoError(t, err)

	respItemArray := make([]ResponseItem, 0)

	for contract, item := range expectedSolcResponse.Contracts {
		respItem := ResponseItem{
			Objectname: objectName(strings.TrimSpace(contract)),
			Filename:   "simpleContract.sol",
		}
		respItem.Contract.Abi = json.RawMessage(item.Abi)
		respItem.Contract.Evm.Bytecode.Object = item.Bin
		respItem.Contract.Evm.Bytecode.LinkReferences = []byte("{}")
		respItemArray = append(respItemArray, respItem)
	}
	expectedResponse := &Response{
		Objects: respItemArray,
		Warning: warning,
		Version: "",
		Error:   "",
	}
	resp, err := EVM("simpleContract.sol", false, "", make(map[string]string), logging.NewNoopLogger())
	if err != nil {
		t.Fatal(err)
	}
	for i := range resp.Objects {
		resp.Objects[i].Contract.Metadata = ""
		resp.Objects[i].Contract.Devdoc = nil
		resp.Objects[i].Contract.MetadataMap = nil
		resp.Objects[i].Contract.Evm.DeployedBytecode.Object = ""
		resp.Objects[i].Contract.Evm.DeployedBytecode.LinkReferences = nil
	}
	assert.Equal(t, expectedResponse, resp)
}

func TestFaultyContract(t *testing.T) {
	const faultyContractFile = "tests/compilers_fixtures/faultyContract.sol"
	actualOutput, err := exec.Command("solc", "--combined-json", "bin,abi", faultyContractFile).CombinedOutput()
	require.EqualError(t, err, "exit status 1")
	resp, err := EVM(faultyContractFile, false, "", make(map[string]string), logging.NewNoopLogger())
	require.NoError(t, err)
	if err != nil {
		if string(actualOutput) != resp.Error {
			t.Errorf("Expected %v got %v", string(actualOutput), resp.Error)
		}
	}
	output := strings.TrimSpace(string(actualOutput))
	fmt.Println(output)
}

func testContractPath() string {
	baseDir, _ := os.Getwd()
	return filepath.Join(baseDir, "..", "..", "tests", "compilers_fixtures")
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
		if a.Contract.Evm.Bytecode.Object == e.Contract.Evm.Bytecode.Object {
			return true
		}
	}
	return false
}
