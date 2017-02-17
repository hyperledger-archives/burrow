package vm

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"fmt"
)

/* Compiling the Permissions solidity contract at
https://ethereum.github.io/browser-solidity yields:

3fbf7da5 add_role(address,bytes32)
bb37737a has_base(address,int256)
e8145855 has_role(address,bytes32)
28fd0194 rm_role(address,bytes32)
9ea53314 set_base(address,int256,bool)
d69186a6 set_global(int256,bool)
180d26f2 unset_base(address,int256)
*/

func TestPermissionsContract(t *testing.T) {
	registerNativeContracts()
	contract := SNativeContracts()["permissions_contract"]

	assertContractFunction(t, contract, "3fbf7da5",
		"add_role(address,bytes32)")

	assertContractFunction(t, contract, "bb37737a",
		"has_base(address,int256)")

	assertContractFunction(t, contract, "054556ac",
		"has_role(address,bytes32)")

	assertContractFunction(t, contract, "ded3350a",
		"rm_role(address,bytes32)")

	assertContractFunction(t, contract, "c2174d8f",
		"set_base(address,int256,bool)")

	assertContractFunction(t, contract, "85f1522b",
		"set_global(int256,bool)")

	assertContractFunction(t, contract, "73448c99",
		"unset_base(address,int256)")
}

func TestSNativeFuncTemplate(t *testing.T) {
	contract := SNativeContracts()["permissions_contract"]
	function, err := contract.FunctionByName("rm_role")
	if err != nil {
		t.Fatal("Couldn't get function")
	}
	solidity, err := function.Solidity()
	assert.NoError(t, err)
	fmt.Println(solidity)
}

// This test checks that we can generate the SNative contract interface and
// prints it to stdout
func TestSNativeContractTemplate(t *testing.T) {
	contract := SNativeContracts()["permissions_contract"]
	solidity, err := contract.Solidity()
	assert.NoError(t, err)
	fmt.Println(solidity)
}

// Helpers

func assertContractFunction(t *testing.T, contract SNativeContractDescription,
	funcIDHex string, expectedSignature string) {
	function, err := contract.FunctionByID(fourBytesFromHex(t, funcIDHex))
	assert.NoError(t, err,
		"Error retrieving SNativeFunctionDescription with ID %s", funcIDHex)
	if err == nil {
		assert.Equal(t, expectedSignature, function.Signature())
	}
}

func fourBytesFromHex(t *testing.T, hexString string) [4]byte {
	bs, err := hex.DecodeString(hexString)
	assert.NoError(t, err, "Could not decode hex string '%s'", hexString)
	if len(bs) != 4 {
		t.Fatalf("FuncID must be 4 bytes but '%s' is %v bytes", hexString,
			len(bs))
	}
	return firstFourBytes(bs)
}

