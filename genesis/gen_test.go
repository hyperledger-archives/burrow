package genesis

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// set the chain ID
var chainID string = "genesis-file-maker-test"

var genesisFileExpected string = `{
	"genesis_time": "0001-01-01T00:00:00.000Z",
	"chain_id": "genesis-file-maker-test",
	"params": null,
	"accounts": [
		{
			"address": "74417C1BEFB3938B71B22B202050A4C6591FFCF6",
			"amount": 9999999999,
			"name": "genesis-file-maker-test_developer_000",
			"permissions": {
				"base": {
					"perms": 14430,
					"set": 16383
				},
				"roles": []
			}
		},
		{
			"address": "0C9DAEA4046491A661FCE0B41B0CAA2AD3415268",
			"amount": 99999999999999,
			"name": "genesis-file-maker-test_full_000",
			"permissions": {
				"base": {
					"perms": 16383,
					"set": 16383
				},
				"roles": []
			}
		},
		{
			"address": "E1BD50A1B90A15861F5CF0F182D291F556B21A86",
			"amount": 9999999999,
			"name": "genesis-file-maker-test_participant_000",
			"permissions": {
				"base": {
					"perms": 2118,
					"set": 16383
				},
				"roles": []
			}
		},
		{
			"address": "A6C8E2DE652DB8ADB4036293DC21F8FE389D77C2",
			"amount": 9999999999,
			"name": "genesis-file-maker-test_root_000",
			"permissions": {
				"base": {
					"perms": 16383,
					"set": 16383
				},
				"roles": []
			}
		},
		{
			"address": "E96CB7910001320B6F1E2266A8431D5E98FF0183",
			"amount": 9999999999,
			"name": "genesis-file-maker-test_validator_000",
			"permissions": {
				"base": {
					"perms": 32,
					"set": 16383
				},
				"roles": []
			}
		}
	],
	"validators": [
		{
			"pub_key": [
				1,
				"238E1A77CC7CDCD13F4D77841F1FE4A46A77DB691EC89718CD0D4CB3409F61D2"
			],
			"amount": 9999999999,
			"name": "genesis-file-maker-test_full_000",
			"unbond_to": [
				{
					"address": "0C9DAEA4046491A661FCE0B41B0CAA2AD3415268",
					"amount": 9999999999
				}
			]
		},
		{
			"pub_key": [
				1,
				"7F53D78C526F96C87ACBD0D2B9DB2E9FC176981623D26B1DB1CF59748EE9F4CF"
			],
			"amount": 9999999998,
			"name": "genesis-file-maker-test_validator_000",
			"unbond_to": [
				{
					"address": "E96CB7910001320B6F1E2266A8431D5E98FF0183",
					"amount": 9999999998
				}
			]
		}
	]
}`

var accountsCSV string = `F0BD5CE45D306D61C9AB73CE5268C2B59D52CAF7127EF0E3B65523302254350A,9999999999,genesis-file-maker-test_developer_000,14430,16383
238E1A77CC7CDCD13F4D77841F1FE4A46A77DB691EC89718CD0D4CB3409F61D2,99999999999999,genesis-file-maker-test_full_000,16383,16383
E37A655E560D53721C9BB06BA742398323504DFE2EB2C67E71F8D16E71E0471B,9999999999,genesis-file-maker-test_participant_000,2118,16383
EC0E38CC8308EC9E720EE839242A7BC5C781D1F852E962FAC5A8E0599CE5B224,9999999999,genesis-file-maker-test_root_000,16383,16383
7F53D78C526F96C87ACBD0D2B9DB2E9FC176981623D26B1DB1CF59748EE9F4CF,9999999999,genesis-file-maker-test_validator_000,32,16383`

var validatorsCSV string = `238E1A77CC7CDCD13F4D77841F1FE4A46A77DB691EC89718CD0D4CB3409F61D2,9999999999,genesis-file-maker-test_full_000,16383,16383
7F53D78C526F96C87ACBD0D2B9DB2E9FC176981623D26B1DB1CF59748EE9F4CF,9999999998,genesis-file-maker-test_validator_000,32,16383`

func TestKnownCSV(t *testing.T) {
	// make temp dir
	dir, err := ioutil.TempDir(os.TempDir(), "genesis-file-maker-test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		//cleanup
		os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}

	}()

	// set the filepaths to be written to
	accountsCSVpath := filepath.Join(dir, "accounts.csv")
	validatorsCSVpath := filepath.Join(dir, "validators.csv")

	// write the accounts.csv
	if err := ioutil.WriteFile(accountsCSVpath, []byte(accountsCSV), 0600); err != nil {
		t.Fatal(err)
	}

	// write the validators.csv
	if err := ioutil.WriteFile(validatorsCSVpath, []byte(validatorsCSV), 0600); err != nil {
		t.Fatal(err)
	}

	// create the genesis file
	// NOTE: [ben] set time to zero time, "genesis_time": "0001-01-01T00:00:00.000Z"
	genesisFileWritten, err := generateKnownWithTime(chainID, accountsCSVpath, validatorsCSVpath, time.Time{})
	if err != nil {
		t.Fatal(err)
	}

	// compare
	if !bytes.Equal([]byte(genesisFileExpected), []byte(genesisFileWritten)) {
		t.Fatalf("Bad genesis file: got (%s), expected (%s)", genesisFileWritten, genesisFileExpected)
	}

}
