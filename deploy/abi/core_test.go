package abi

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	pm "github.com/hyperledger/burrow/deploy/def"
	"github.com/stretchr/testify/require"
	"github.com/tmthrgd/go-hex"
)

//To Test:
//Bools, Arrays, Addresses, Hashes
//Test Packing different things
//After that, should be good to go

func TestPacker(t *testing.T) {
	for _, test := range []struct {
		ABI            string
		args           []string
		name           string
		expectedOutput []byte
	}{
		{
			`[{"constant":false,"inputs":[{"name":"","type":"uint256"}],"name":"UInt","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"1"},
			"UInt",
			pad([]byte{1}, 32, true),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"int256"}],"name":"Int","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"-1"},
			"Int",
			[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"bool"}],"name":"Bool","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"true"},
			"Bool",
			pad([]byte{1}, 32, true),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"string"}],"name":"String","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"marmots"},
			"String",
			append(hexToBytes(t, "00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000007"), pad([]byte("marmots"), 32, false)...),
		},
		{
			`[{"constant":false,"inputs":[{"name":"x","type":"bytes32"}],"name":"Bytes32","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"marmatoshi"},
			"Bytes32",
			pad([]byte("marmatoshi"), 32, false),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"uint8"}],"name":"UInt8","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"1"},
			"UInt8",
			pad([]byte{1}, 32, true),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"int8"}],"name":"Int8","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"-1"},
			"Int8",
			[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"uint256"},{"name":"","type":"uint256"}],"name":"multiPackUInts","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"1", "1"},
			"multiPackUInts",
			append(pad([]byte{1}, 32, true), pad([]byte{1}, 32, true)...),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"bool"},{"name":"","type":"bool"}],"name":"multiPackBools","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"true", "false"},
			"multiPackBools",
			append(pad([]byte{1}, 32, true), pad([]byte{0}, 32, true)...),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"int256"},{"name":"","type":"int256"}],"name":"multiPackInts","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"-1", "-1"},
			"multiPackInts",
			[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		},

		{
			`[{"constant":false,"inputs":[{"name":"","type":"string"},{"name":"","type":"string"}],"name":"multiPackStrings","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"hello", "world"},
			"multiPackStrings",
			append(
				hexToBytes(t, "000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000005"),
				append(pad([]byte("hello"), 32, false),
					append(hexToBytes(t, "0000000000000000000000000000000000000000000000000000000000000005"),
						pad([]byte("world"), 32, false)...)...)...,
			),
		},
		{
			`[{"constant":false,"inputs":[],"name":"arrayOfBytes32Pack","inputs":[{"name":"","type":"bytes32[3]"}],"payable":false,"type":"function"}]`,
			[]string{`[den,of,marmots]`},
			"arrayOfBytes32Pack",
			append(
				pad([]byte("den"), 32, false),
				append(pad([]byte("of"), 32, false), pad([]byte("marmots"), 32, false)...)...,
			),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"uint256[3]"}],"name":"arrayOfUIntsPack","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"[1,2,3]"},
			"arrayOfUIntsPack",
			hexToBytes(t, "000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003"),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"int256[3]"}],"name":"arrayOfIntsPack","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"[-1,-2,-3]"},
			"arrayOfIntsPack",
			[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 254, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 253},
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"bool[2]"}],"name":"arrayOfBoolsPack","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"[true,false]"},
			"arrayOfBoolsPack",
			append(pad([]byte{1}, 32, true), pad([]byte{0}, 32, true)...),
		},
	} {
		t.Log(test.args)
		fmt.Println(test.name)
		/*abiStruct, err := JSON(strings.NewReader(test.ABI))
		if err != nil {
			t.Errorf("Incorrect ABI: ", err)
		}*/
		if output, err := Packer(test.ABI, test.name, test.args...); err != nil {
			t.Error("Unexpected error in ", test.name, ": ", err)
		} else {
			if !bytes.Equal(output[4:], test.expectedOutput) {
				t.Errorf("Incorrect output,\n\t expected %v,\n\t got %v", test.expectedOutput, output[4:])
			}
		}
	}
}

func TestUnpacker(t *testing.T) {
	for _, test := range []struct {
		abi            string
		packed         []byte
		name           string
		expectedOutput []pm.Variable
	}{
		{
			`[{"constant":true,"inputs":[],"name":"String","outputs":[{"name":"","type":"string"}],"payable":false,"type":"function"}]`,
			append(pad(hexToBytes(t, "0000000000000000000000000000000000000000000000000000000000000020"), 32, true), append(pad(hexToBytes(t, "0000000000000000000000000000000000000000000000000000000000000005"), 32, true), pad([]byte("Hello"), 32, false)...)...),
			"String",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "Hello",
				},
			},
		},
		{
			`[{"constant":true,"inputs":[],"name":"UInt","outputs":[{"name":"","type":"uint256"}],"payable":false,"type":"function"}]`,
			hexToBytes(t, "0000000000000000000000000000000000000000000000000000000000000001"),
			"UInt",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "1",
				},
			},
		},
		{
			`[{"constant":false,"inputs":[],"name":"Int","outputs":[{"name":"retVal","type":"int256"}],"payable":false,"type":"function"}]`,
			[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
			"Int",
			[]pm.Variable{
				{
					Name:  "retVal",
					Value: "-1",
				},
			},
		},
		{
			`[{"constant":true,"inputs":[],"name":"Bool","outputs":[{"name":"","type":"bool"}],"payable":false,"type":"function"}]`,
			hexToBytes(t, "0000000000000000000000000000000000000000000000000000000000000001"),
			"Bool",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "true",
				},
			},
		},
		{
			`[{"constant":true,"inputs":[],"name":"Address","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"}]`,
			hexToBytes(t, "0000000000000000000000001040E6521541DAB4E7EE57F21226DD17CE9F0FB7"),
			"Address",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "1040E6521541DAB4E7EE57F21226DD17CE9F0FB7",
				},
			},
		},
		{
			`[{"constant":false,"inputs":[],"name":"Bytes32","outputs":[{"name":"retBytes","type":"bytes32"}],"payable":false,"type":"function"}]`,
			pad([]byte("marmatoshi"), 32, true),
			"Bytes32",
			[]pm.Variable{
				{
					Name:  "retBytes",
					Value: "marmatoshi",
				},
			},
		},
		{
			`[{"constant":false,"inputs":[],"name":"multiReturnUIntInt","outputs":[{"name":"","type":"uint256"},{"name":"","type":"int256"}],"payable":false,"type":"function"}]`,
			append(
				hexToBytes(t, "0000000000000000000000000000000000000000000000000000000000000001"),
				[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}...,
			),
			"multiReturnUIntInt",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "1",
				},
				{
					Name:  "1",
					Value: "-1",
				},
			},
		},
		{
			`[{"constant":false,"inputs":[],"name":"multiReturnMixed","outputs":[{"name":"","type":"string"},{"name":"","type":"uint256"}],"payable":false,"type":"function"}]`,
			append(
				hexToBytes(t, "00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000001"),
				append(hexToBytes(t, "0000000000000000000000000000000000000000000000000000000000000005"), pad([]byte("Hello"), 32, false)...)...,
			),
			"multiReturnMixed",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "Hello",
				},
				{
					Name:  "1",
					Value: "1",
				},
			},
		},
		{
			`[{"constant":false,"inputs":[],"name":"multiPackBytes32","outputs":[{"name":"","type":"bytes32"},{"name":"","type":"bytes32"},{"name":"","type":"bytes32"}],"payable":false,"type":"function"}]`,
			append(
				pad([]byte("den"), 32, true),
				append(pad([]byte("of"), 32, true), pad([]byte("marmots"), 32, true)...)...,
			),
			"multiPackBytes32",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "den",
				},
				{
					Name:  "1",
					Value: "of",
				},
				{
					Name:  "2",
					Value: "marmots",
				},
			},
		},
		{
			`[{"constant":false,"inputs":[],"name":"arrayReturnBytes32","outputs":[{"name":"","type":"bytes32[3]"}],"payable":false,"type":"function"}]`,
			append(
				pad([]byte("den"), 32, true),
				append(pad([]byte("of"), 32, true), pad([]byte("marmots"), 32, true)...)...,
			),
			"arrayReturnBytes32",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "[den,of,marmots]",
				},
			},
		},
		{
			`[{"constant":false,"inputs":[],"name":"arrayReturnUInt","outputs":[{"name":"","type":"uint256[3]"}],"payable":false,"type":"function"}]`,
			hexToBytes(t, "000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003"),
			"arrayReturnUInt",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "[1,2,3]",
				},
			},
		},
		{
			`[{"constant":false,"inputs":[],"name":"arrayReturnInt","outputs":[{"name":"","type":"int256[2]"}],"payable":false,"type":"function"}]`,
			[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 253, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 254},
			"arrayReturnInt",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "[-3,-2]",
				},
			},
		},
	} {
		//t.Log(test.name)
		t.Log(test.packed)
		output, err := Unpacker(test.abi, test.name, test.packed)
		if err != nil {
			t.Errorf("Unpacker failed: %v", err)
		}
		for i, expectedOutput := range test.expectedOutput {

			if output[i].Name != expectedOutput.Name {
				t.Errorf("Unpacker failed: Incorrect Name, got %v expected %v", output[i].Name, expectedOutput.Name)
			}
			//t.Log("Test: ", output[i].Value)
			//t.Log("Test: ", expectedOutput.Value)
			if strings.Compare(output[i].Value, expectedOutput.Value) != 0 {
				t.Errorf("Unpacker failed: Incorrect value, got %v expected %v", output[i].Value, expectedOutput.Value)
			}
		}
	}
}

func hexToBytes(t testing.TB, hexString string) []byte {
	bs, err := hex.DecodeString(hexString)
	require.NoError(t, err)
	return bs
}
