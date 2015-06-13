package test

import (
	edb "github.com/eris-ltd/erisdb/erisdb"
	ep "github.com/eris-ltd/erisdb/erisdb/pipe"
	"github.com/tendermint/tendermint/account"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
)

var testDataJson = `{
  "chain_data": {
    "priv_validator": {
      "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
      "pub_key": [1, "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"],
      "priv_key": [1, "6B72D45EB65F619F11CE580C8CAED9E0BADC774E9C9C334687A65DCBAD2C4151CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"],
      "last_height": 0,
      "last_round": 0,
      "last_step": 0
    },
    "genesis": {
      "chain_id": "my_tests",
      "accounts": [
        {
          "address": "F81CB9ED0A868BD961C4F5BBC0E39B763B89FCB6",
          "amount": 690000000000
        },
        {
          "address": "0000000000000000000000000000000000000002",
          "amount": 565000000000
        },
        {
          "address": "9E54C9ECA9A3FD5D4496696818DA17A9E17F69DA",
          "amount": 525000000000
        },
        {
          "address": "0000000000000000000000000000000000000004",
          "amount": 110000000000
        },
        {
          "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
          "amount": 110000000000
        }

      ],
      "validators": [
        {
          "pub_key": [1, "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"],
          "amount": 5000000000,
          "unbond_to": [
            {
              "address": "93E243AC8A01F723DE353A4FA1ED911529CCB6E5",
              "amount": 5000000000
            }
          ]
        }
      ]
    }
  },
  "input" : {
    "account_address": "9FC1ECFCAE2A554D4D1A000D0D80F748E66359E3",
    "storage_address": "",
    "tx_create" : {
      "address": "",
      "priv_key": "6B72D45EB65F619F11CE580C8CAED9E0BADC774E9C9C334687A65DCBAD2C4151CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906",
      "data": "5B33600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF021916908302179055505B6102828061003B6000396000F3006000357C01000000000000000000000000000000000000000000000000000000009004806337F428411461004557806340C10F191461005A578063D0679D341461006E57005B610050600435610244565B8060005260206000F35B610068600435602435610082565B60006000F35B61007C600435602435610123565B60006000F35B600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1673FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF163373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1614156100DD576100E2565B61011F565B80600160005060008473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020019081526020016000206000828282505401925050819055505B5050565B80600160005060003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF168152602001908152602001600020600050541061015E57610163565B610240565B80600160005060003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200190815260200160002060008282825054039250508190555080600160005060008473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020019081526020016000206000828282505401925050819055507F93EB3C629EB575EDAF0252E4F9FC0C5CCADA50496F8C1D32F0F93A65A8257EB560003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020018373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020018281526020016000A15B5050565B6000600160005060008373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200190815260200160002060005054905061027D565B91905056",
      "fee" : 1000,
      "gas_limit": 1000
    },
    "tx": {
      "address": "9FC1ECFCAE2A554D4D1A000D0D80F748E66359E3",
      "priv_key": "6B72D45EB65F619F11CE580C8CAED9E0BADC774E9C9C334687A65DCBAD2C4151CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906",
      "data": "",
      "fee" : 1000,
      "gas_limit": 1000
    },
    "call_code": {
      "code": "5B33600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF021916908302179055505B6102828061003B6000396000F3006000357C01000000000000000000000000000000000000000000000000000000009004806337F428411461004557806340C10F191461005A578063D0679D341461006E57005B610050600435610244565B8060005260206000F35B610068600435602435610082565B60006000F35B61007C600435602435610123565B60006000F35B600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1673FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF163373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1614156100DD576100E2565B61011F565B80600160005060008473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020019081526020016000206000828282505401925050819055505B5050565B80600160005060003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF168152602001908152602001600020600050541061015E57610163565B610240565B80600160005060003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200190815260200160002060008282825054039250508190555080600160005060008473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020019081526020016000206000828282505401925050819055507F93EB3C629EB575EDAF0252E4F9FC0C5CCADA50496F8C1D32F0F93A65A8257EB560003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020018373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020018281526020016000A15B5050565B6000600160005060008373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200190815260200160002060005054905061027D565B91905056",
      "data": ""
    },
    "block_range": {
      "min": 0,
      "max": 0
    }
  },
  "output": {
    "consensus_state": {
      "height": 1,
      "round": 0,
      "step": 1,
      "start_time": "",
      "commit_time": "0001-01-01 00:00:00 +0000 UTC",
      "validators": [{
        "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
        "pub_key": [1, "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"],
        "bond_height": 0,
        "unbond_height": 0,
        "last_commit_height": 0,
        "voting_power": 5000000000,
        "accum": 0
      }],
      "proposal": null
    },
    "validators": {
      "block_height": 0,
      "bonded_validators": [{
        "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
        "pub_key": [1, "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"],
        "bond_height": 0,
        "unbond_height": 0,
        "last_commit_height": 0,
        "voting_power": 5000000000,
        "accum": 0
      }],
      "unbonding_validators": []
    },
    "network_info": {
      "moniker": "anothertester",
      "listening": false,
      "listeners": [],
      "peers": []
    },
    "moniker": {
      "moniker": "anothertester"
    },
    "listening": {
      "listening": false
    },
    "listeners": {
      "listeners": []
    },
    "peers": [],
    "tx_create_receipt": {
      "tx_hash": "C1C84BCD4CCA6D6132D1E702FA4A7618DBCDB52F",
      "creates_contract": 1,
      "contract_addr": "9FC1ECFCAE2A554D4D1A000D0D80F748E66359E3"
    },
    "tx_receipt": {
      "tx_hash": "A40873D4C7136F6D79476A035D4265781FC20A3B",
      "creates_contract": 0,
      "contract_addr": ""
    },
    "unconfirmed_txs": {
      "txs": [
        [ 2, {
          "input": {
            "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
            "amount": 1000,
            "sequence": 1,
            "signature": [ 1, "39E1D98C2F4F8FC5A98442C55DCC8FCBCE4ADB0F6BAD5C5258CEFE94EFB0315EA9616CC275F97E4D04C5A8FD08D73B84A28B7CFEAEE98F4A37E2F2BCA1830907" ],
            "pub_key": [1, "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"]
          },
          "address": "",
          "gas_limit": 1000,
          "fee": 1000,
          "data": "5B33600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF021916908302179055505B6102828061003B6000396000F3006000357C01000000000000000000000000000000000000000000000000000000009004806337F428411461004557806340C10F191461005A578063D0679D341461006E57005B610050600435610244565B8060005260206000F35B610068600435602435610082565B60006000F35B61007C600435602435610123565B60006000F35B600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1673FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF163373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1614156100DD576100E2565B61011F565B80600160005060008473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020019081526020016000206000828282505401925050819055505B5050565B80600160005060003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF168152602001908152602001600020600050541061015E57610163565B610240565B80600160005060003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200190815260200160002060008282825054039250508190555080600160005060008473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020019081526020016000206000828282505401925050819055507F93EB3C629EB575EDAF0252E4F9FC0C5CCADA50496F8C1D32F0F93A65A8257EB560003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020018373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020018281526020016000A15B5050565B6000600160005060008373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200190815260200160002060005054905061027D565B91905056"
        } ],
        [ 2, {
          "input": {
            "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
            "amount": 1000,
            "sequence": 3,
            "signature": [1, "8D84089EC1140C5AF474DB7E764E937D9C6309BA0AD7BCC56108A2075504005AE2EE1AD71C3D414F9D793D2BFAD77C9572D9494736E6F3D1A62D17DF4A01090D" ],
            "pub_key": null
          },
          "address": "9FC1ECFCAE2A554D4D1A000D0D80F748E66359E3",
          "gas_limit": 1000,
          "fee": 1000,
          "data": ""
        } ]
      ]},
    "call_code": {
      "return": "6000357c01000000000000000000000000000000000000000000000000000000009004806337f428411461004557806340c10f191461005a578063d0679d341461006e57005b610050600435610244565b8060005260206000f35b610068600435602435610082565b60006000f35b61007c600435602435610123565b60006000f35b600060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614156100dd576100e2565b61011f565b80600160005060008473ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828282505401925050819055505b5050565b80600160005060003373ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600050541061015e57610163565b610240565b80600160005060003373ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282825054039250508190555080600160005060008473ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828282505401925050819055507f93eb3c629eb575edaf0252e4f9fc0c5ccada50496f8c1d32f0f93a65a8257eb560003373ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1681526020018281526020016000a15b5050565b6000600160005060008373ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060005054905061027d565b91905056",
      "gas_used": 0
    },
    "accounts": {
      "accounts": [
        {
          "address": "0000000000000000000000000000000000000002",
          "pub_key": null,
          "sequence": 0,
          "balance": 565000000000,
          "code": "",
          "storage_root": ""
        },
        {
          "address": "0000000000000000000000000000000000000004",
          "pub_key": null,
          "sequence": 0,
          "balance": 110000000000,
          "code": "",
          "storage_root": ""
        },
        {
          "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
          "pub_key": null,
          "sequence": 0,
          "balance": 110000000000,
          "code": "",
          "storage_root": ""
        },
        {
          "address": "9E54C9ECA9A3FD5D4496696818DA17A9E17F69DA",
          "pub_key": null,
          "sequence": 0,
          "balance": 525000000000,
          "code": "",
          "storage_root": ""
        },
        {
          "address": "F81CB9ED0A868BD961C4F5BBC0E39B763B89FCB6",
          "pub_key": null,
          "sequence": 0,
          "balance": 690000000000,
          "code": "",
          "storage_root": ""
        }

      ]
    },
    "account": {
      "address": "9FC1ECFCAE2A554D4D1A000D0D80F748E66359E3",
      "pub_key": null,
      "sequence": 0,
      "balance": 0,
      "code": "",
      "storage_root": ""
    },
    "storage": {
      "storage_root": "",
      "storage_items": []
    },
    "storage_at": {},
    "blockchain_info": {
      "chain_id": "my_tests",
      "genesis_hash": "DA4F4DC4A54620F1E0AA1213631C4DC2957B7415E3F8C066C30009BC57C4E5FC",
      "latest_block_height": 0,
      "latest_block": null
    },
    "chain_id": {
      "chain_id": "my_tests"
    },
    "genesis_hash": {
      "hash": "DA4F4DC4A54620F1E0AA1213631C4DC2957B7415E3F8C066C30009BC57C4E5FC"
    },
    "latest_block_height": {
      "height": 0
    },
    "block": null,
    "blocks": {
      "min_height": 0,
      "max_height": 0,
      "block_metas": []
    }
  }
}`

var serverDuration uint = 100

type (
	BlockRange struct {
		Min int `json:"min"`
		Max int `json:"max"`
	}
)

type (
	ChainData struct {
		PrivValidator *state.PrivValidator `json:"priv_validator"`
		Genesis       *state.GenesisDoc    `json:"genesis"`
	}

	Input struct {
		AccountAddress string             `json:"account_address"`
		StorageAddress string             `json:"storage_address"`
		TxCreate       *edb.TransactParam `json:"tx_create"`
		Tx             *edb.TransactParam `json:"tx"`
		CallCode       *edb.CallCodeParam `json:"call_code"`
		BlockRange     *BlockRange        `json:"block_range"`
	}

	Output struct {
		ConsensusState    *ep.ConsensusState    `json:"consensus_state"`
		Validators        *ep.ValidatorList     `json:"validators"`
		NetworkInfo       *ep.NetworkInfo       `json:"network_info"`
		Moniker           *ep.Moniker           `json:"moniker"`
		Listening         *ep.Listening         `json:"listening"`
		Listeners         *ep.Listeners         `json:"listeners"`
		Peers             []*ep.Peer            `json:"peers"`
		TxCreateReceipt   *ep.Receipt           `json:"tx_create_receipt"`
		TxReceipt         *ep.Receipt           `json:"tx_receipt"`
		UnconfirmedTxs    *ep.UnconfirmedTxs    `json:"unconfirmed_txs"`
		CallCode          *ep.Call              `json:"call_code"`
		Accounts          *ep.AccountList       `json:"accounts"`
		Account           *account.Account      `json:"account"`
		Storage           *ep.Storage           `json:"storage"`
		StorageAt         *ep.StorageItem       `json:"storage_at"`
		BlockchainInfo    *ep.BlockchainInfo    `json:"blockchain_info"`
		ChainId           *ep.ChainId           `json:"chain_id"`
		GenesisHash       *ep.GenesisHash       `json:"genesis_hash"`
		LatestBlockHeight *ep.LatestBlockHeight `json:"latest_block_height"`
		Block             *types.Block          `json:"block"`
		Blocks            *ep.Blocks            `json:"blocks"`
	}

	TestData struct {
		ChainData *ChainData `json:"chain_data"`
		Input     *Input     `json:"input"`
		Output    *Output    `json:"output"`
	}
)

func LoadTestData() *TestData {
	codec := edb.NewTCodec()
	testData := &TestData{}
	err := codec.DecodeBytes(testData, []byte(testDataJson))
	// TODO for now.
	if err != nil {
		panic(err)
	}
	return testData
}
