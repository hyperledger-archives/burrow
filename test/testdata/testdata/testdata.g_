package testdata

import (

	account    "github.com/eris-ltd/eris-db/account"
	core_types "github.com/eris-ltd/eris-db/core/types"
	event      "github.com/eris-ltd/eris-db/event"
	rpc_v0     "github.com/eris-ltd/eris-db/rpc/v0"
	stypes     "github.com/eris-ltd/eris-db/manager/eris-mint/state/types"
	types      "github.com/eris-ltd/eris-db/txs"

	mintTypes "github.com/tendermint/tendermint/types"
)

var testDataJson = `{
  "chain_data": {
    "priv_validator": {
      "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
      "pub_key": [
        1,
        "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"
      ],
      "priv_key": [
        1,
        "6B72D45EB65F619F11CE580C8CAED9E0BADC774E9C9C334687A65DCBAD2C4151CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"
      ],
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
          "pub_key": "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906",
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
  "GetAccount": {
    "input": {
      "address": "9FC1ECFCAE2A554D4D1A000D0D80F748E66359E3"
    },
    "output": {
      "address": "9FC1ECFCAE2A554D4D1A000D0D80F748E66359E3",
      "pub_key": null,
      "sequence": 0,
      "balance": 0,
      "code": "",
      "storage_root": "",
      "permissions": {
        "base": {
          "perms": 0,
          "set": 0
        },
        "roles": []
      }
    }
  },
  "GetAccounts": {
    "input": {
      "filters": []
    },
    "output": {
      "accounts": [
        {
	      "address": "0000000000000000000000000000000000000000",
	      "pub_key": null,
	      "sequence": 0,
	      "balance": 1337,
	      "code": "",
	      "storage_root": "",
	      "permissions": {
	        "base": {
	          "perms": 2302,
	          "set": 16383
	        },
	        "roles": []
	      }
	    },
        {
          "address": "0000000000000000000000000000000000000002",
          "pub_key": null,
          "sequence": 0,
          "balance": 565000000000,
          "code": "",
          "storage_root": "",
          "permissions": {
            "base": {
              "perms": 0,
              "set": 0
            },
            "roles": []
          }
        },
        {
          "address": "0000000000000000000000000000000000000004",
          "pub_key": null,
          "sequence": 0,
          "balance": 110000000000,
          "code": "",
          "storage_root": "",
          "permissions": {
            "base": {
              "perms": 0,
              "set": 0
            },
            "roles": []
          }
        },
        {
          "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
          "pub_key": null,
          "sequence": 0,
          "balance": 110000000000,
          "code": "",
          "storage_root": "",
          "permissions": {
            "base": {
              "perms": 0,
              "set": 0
            },
            "roles": []
          }
        },
        {
          "address": "9E54C9ECA9A3FD5D4496696818DA17A9E17F69DA",
          "pub_key": null,
          "sequence": 0,
          "balance": 525000000000,
          "code": "",
          "storage_root": "",
          "permissions": {
            "base": {
              "perms": 0,
              "set": 0
            },
            "roles": []
          }
        },
        {
          "address": "F81CB9ED0A868BD961C4F5BBC0E39B763B89FCB6",
          "pub_key": null,
          "sequence": 0,
          "balance": 690000000000,
          "code": "",
          "storage_root": "",
          "permissions": {
            "base": {
              "perms": 0,
              "set": 0
            },
            "roles": []
          }
        }
      ]
    }
  },
  "GetStorage": {
    "input": {
      "address": "9FC1ECFCAE2A554D4D1A000D0D80F748E66359E3"
    },
    "output": {
      "storage_root": "",
      "storage_items": []
    }
  },
  "GetStorageAt": {
    "input": {
      "address": "9FC1ECFCAE2A554D4D1A000D0D80F748E66359E3",
      "key": "00"
    },
    "output": {
      "key": "00",
      "value": ""
    }
  },
  "GenPrivAccount": {
    "output": {
      "address": "",
      "pub_key": [
        1,
        "0000000000000000000000000000000000000000000000000000000000000000"
      ],
      "priv_key": [
        1,
        "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
      ]
    }
  },
  "GetBlockchainInfo": {
    "output": {
      "chain_id": "my_tests",
      "genesis_hash": "0A8C453DB67BE52D32F9451212E8CE0E172AE56C",
      "latest_block_height": 0,
      "latest_block": null
    }
  },
  "GetChainId": {
    "output": {
      "chain_id": "my_tests"
    }
  },
  "GetGenesisHash": {
    "output": {
      "hash": "0A8C453DB67BE52D32F9451212E8CE0E172AE56C"
    }
  },
  "GetLatestBlockHeight": {
    "output": {
      "height": 0
    }
  },
  "GetLatestBlock": {
    "output": {}
  },
  "GetBlock": {
    "input": {"height": 0},
    "output": null
  },
  "GetBlocks": {
    "input": {
      "filters": []
    },
    "output": {
      "min_height": 0,
      "max_height": 0,
      "block_metas": []
    }
  },
  "GetConsensusState": {
    "output": {
      "height": 1,
      "round": 0,
      "step": 1,
      "start_time": "",
      "commit_time": "0001-01-01 00:00:00 +0000 UTC",
      "validators": [
        {
          "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
          "pub_key": [1, "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"],
          "bond_height": 0,
          "unbond_height": 0,
          "last_commit_height": 0,
          "voting_power": 5000000000,
          "accum": 0
        }
      ],
      "proposal": null
    }
  },
  "GetValidators": {
    "output": {
      "block_height": 0,
      "bonded_validators": [
        {
          "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
          "pub_key": [1, "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"],
          "bond_height": 0,
          "unbond_height": 0,
          "last_commit_height": 0,
          "voting_power": 5000000000,
          "accum": 0
        }
      ],
      "unbonding_validators": []
    }
  },
  "GetNetworkInfo": {
    "output": {
      "client_version": "0.5.2",
      "moniker": "__MONIKER__",
      "listening": false,
      "listeners": [],
      "peers": []
    }
  },
  "GetClientVersion": {
    "output": {
      "client_version": "0.5.2"
    }
  },
  "GetMoniker": {
    "output": {
      "moniker": "__MONIKER__"
    }
  },
  "IsListening": {
    "output": {
      "listening": false
    }
  },
  "GetListeners": {
    "output": {
      "listeners": []
    }
  },
  "GetPeers": {
    "output": []
  },
  "GetPeer": {
    "input": {"address": "127.0.0.1:30000"},
    "output": {
      "is_outbound": false,
      "node_info": null
    }
  },
  "Transact": {
    "input": {
      "address": "9FC1ECFCAE2A554D4D1A000D0D80F748E66359E3",
      "priv_key": "6B72D45EB65F619F11CE580C8CAED9E0BADC774E9C9C334687A65DCBAD2C4151CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906",
      "data": "",
      "fee": 0,
      "gas_limit": 1000000
    },
    "output": {
      "tx_hash": "240E5BDCC0E4F7C1F29A66CA20E3F7A0D6F7EF51",
      "creates_contract": 0,
      "contract_addr": ""
    }
  },
  "TransactCreate": {
    "input": {
      "address": "",
      "priv_key": "6B72D45EB65F619F11CE580C8CAED9E0BADC774E9C9C334687A65DCBAD2C4151CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906",
      "data": "5B33600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF021916908302179055505B6102828061003B6000396000F3006000357C01000000000000000000000000000000000000000000000000000000009004806337F428411461004557806340C10F191461005A578063D0679D341461006E57005B610050600435610244565B8060005260206000F35B610068600435602435610082565B60006000F35B61007C600435602435610123565B60006000F35B600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1673FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF163373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1614156100DD576100E2565B61011F565B80600160005060008473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020019081526020016000206000828282505401925050819055505B5050565B80600160005060003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF168152602001908152602001600020600050541061015E57610163565B610240565B80600160005060003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200190815260200160002060008282825054039250508190555080600160005060008473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020019081526020016000206000828282505401925050819055507F93EB3C629EB575EDAF0252E4F9FC0C5CCADA50496F8C1D32F0F93A65A8257EB560003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020018373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020018281526020016000A15B5050565B6000600160005060008373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200190815260200160002060005054905061027D565B91905056",
      "fee": 0,
      "gas_limit": 1000000
    },
    "output": {
      "tx_hash": "BD5D35871770DB04726843A4C07A26CDE69EB860",
      "creates_contract": 1,
      "contract_addr": "576439CD5C22EB6F3AE1AC1EC5101C5CE1E120D8"
    }
  },
  "GetUnconfirmedTxs": {
    "output": {
      "txs": [
        [
          2,
          {
            "input": {
              "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
              "amount": 1,
              "sequence": 1,
              "signature": [
                1,
                "2FE1C5EA3B0A05560073D7BF145C0997803113D27618CBCD71985806255E6492C7DC574AF373D3807068164AF4FE51D8CDA7DCC995E088375B83AEA3F8F6F204"
              ],
              "pub_key": [
                1,
                "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"
              ]
            },
            "address": "",
            "gas_limit": 1000000,
            "fee": 0,
            "data": "5B33600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF021916908302179055505B6102828061003B6000396000F3006000357C01000000000000000000000000000000000000000000000000000000009004806337F428411461004557806340C10F191461005A578063D0679D341461006E57005B610050600435610244565B8060005260206000F35B610068600435602435610082565B60006000F35B61007C600435602435610123565B60006000F35B600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1673FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF163373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1614156100DD576100E2565B61011F565B80600160005060008473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020019081526020016000206000828282505401925050819055505B5050565B80600160005060003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF168152602001908152602001600020600050541061015E57610163565B610240565B80600160005060003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200190815260200160002060008282825054039250508190555080600160005060008473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020019081526020016000206000828282505401925050819055507F93EB3C629EB575EDAF0252E4F9FC0C5CCADA50496F8C1D32F0F93A65A8257EB560003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020018373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020018281526020016000A15B5050565B6000600160005060008373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200190815260200160002060005054905061027D565B91905056"
          }
        ],
        [
          2,
          {
            "input": {
              "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
              "amount": 1,
              "sequence": 3,
              "signature": [
                1,
                "425A4D50350EEB597C48F82924E83F24640F9ECB3886A2B85D0073911AE02FC06F3D0FD480D59140B1D2DA669A9BD0227B31026EF3E0AAD534DCF50784984B01"
              ],
              "pub_key": null
            },
            "address": "9FC1ECFCAE2A554D4D1A000D0D80F748E66359E3",
            "gas_limit": 1000000,
            "fee": 0,
            "data": ""
          }
        ]
      ]
    }
  },
  "CallCode": {
    "input": {
      "from": "DEADBEEF",
      "code": "5B33600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF021916908302179055505B6102828061003B6000396000F3006000357C01000000000000000000000000000000000000000000000000000000009004806337F428411461004557806340C10F191461005A578063D0679D341461006E57005B610050600435610244565B8060005260206000F35B610068600435602435610082565B60006000F35B61007C600435602435610123565B60006000F35B600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1673FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF163373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1614156100DD576100E2565B61011F565B80600160005060008473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020019081526020016000206000828282505401925050819055505B5050565B80600160005060003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF168152602001908152602001600020600050541061015E57610163565B610240565B80600160005060003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200190815260200160002060008282825054039250508190555080600160005060008473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020019081526020016000206000828282505401925050819055507F93EB3C629EB575EDAF0252E4F9FC0C5CCADA50496F8C1D32F0F93A65A8257EB560003373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020018373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1681526020018281526020016000A15B5050565B6000600160005060008373FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200190815260200160002060005054905061027D565B91905056",
      "data": ""
    },
    "output": {
      "return": "6000357c01000000000000000000000000000000000000000000000000000000009004806337f428411461004557806340c10f191461005a578063d0679d341461006e57005b610050600435610244565b8060005260206000f35b610068600435602435610082565b60006000f35b61007c600435602435610123565b60006000f35b600060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614156100dd576100e2565b61011f565b80600160005060008473ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828282505401925050819055505b5050565b80600160005060003373ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600050541061015e57610163565b610240565b80600160005060003373ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282825054039250508190555080600160005060008473ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828282505401925050819055507f93eb3c629eb575edaf0252e4f9fc0c5ccada50496f8c1d32f0f93a65a8257eb560003373ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1681526020018281526020016000a15b5050565b6000600160005060008373ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060005054905061027d565b91905056",
      "gas_used": 0
    }
  },
  "Call": {
    "input": {"address": "9FC1ECFCAE2A554D4D1A000D0D80F748E66359E3", "from": "DEADBEEF", "data": ""},
    "output": {
      "return": "6000357c01000000000000000000000000000000000000000000000000000000009004806337f428411461004557806340c10f191461005a578063d0679d341461006e57005b610050600435610244565b8060005260206000f35b610068600435602435610082565b60006000f35b61007c600435602435610123565b60006000f35b600060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614156100dd576100e2565b61011f565b80600160005060008473ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828282505401925050819055505b5050565b80600160005060003373ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600050541061015e57610163565b610240565b80600160005060003373ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282825054039250508190555080600160005060008473ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828282505401925050819055507f93eb3c629eb575edaf0252e4f9fc0c5ccada50496f8c1d32f0f93a65a8257eb560003373ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1681526020018281526020016000a15b5050565b6000600160005060008373ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060005054905061027d565b91905056",
      "gas_used": 0
    }
  },
  "EventSubscribe": {
    "input": {
      "event_id": "testId"
    },
    "output": {
      "sub_id": "1234123412341234123412341234123412341234123412341234123412341234"
    }
  },
  "EventUnsubscribe": {
    "input": {
      "event_sub": "1234123412341234123412341234123412341234123412341234123412341234"
    },
    "output": {
      "result": true
    }
  },
  "EventPoll": {
    "input": {
      "event_sub": "1234123412341234123412341234123412341234123412341234123412341234"
    },
    "output": {
      "events": [
        {
          "address": "0000000000000000000000009FC1ECFCAE2A554D4D1A000D0D80F748E66359E3",
          "topics": [
            "0FC28FCE5E54AC6458756FC24DC51A931CA7AD21440CFCA44933AE774ED5F70C",
            "0000000000000000000000000000000000000000000000000000000000000005",
            "0000000000000000000000000000000000000000000000000000000000000019",
            "000000000000000000000000000000000000000000000000000000000000001E"
          ],
          "data": "41646465642074776F206E756D62657273000000000000000000000000000000",
          "height": 1
        }
      ]
    }
  },
  "TransactNameReg": {
    "input": {
      "priv_key": "6B72D45EB65F619F11CE580C8CAED9E0BADC774E9C9C334687A65DCBAD2C4151CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906",
      "name": "testKey",
      "data": "testValue",
      "amount": 10000,
      "fee": 0
    },
    "output": {
      "tx_hash": "98B0D5162C7CB86FF94BE2C00469107B7CA51CF3",
      "creates_contract": 0,
      "contract_addr": ""
    }
  },
  "GetNameRegEntry": {
    "input": {
      "name": "testKey"
    },
    "output": {
      "name": "testKey",
      "owner": "37236DF251AB70022B1DA351F08A20FB52443E37",
      "data": "testData",
      "expires": 250 }
  },
  "GetNameRegEntries": {
    "input": {
      "filters": []
    },
    "output": {
      "block_height": 1,
      "names":[ {
        "name": "testKey",
        "owner": "37236DF251AB70022B1DA351F08A20FB52443E37",
        "data": "testData",
        "expires": 250
      } ]
    }
  }
}`

var serverDuration uint = 100

type (
	ChainData struct {
		PrivValidator *mintTypes.PrivValidator `json:"priv_validator"`
		Genesis       *stypes.GenesisDoc       `json:"genesis"`
	}

	GetAccountData struct {
		Input  *rpc_v0.AddressParam `json:"input"`
		Output *account.Account         `json:"output"`
	}

	GetAccountsData struct {
		Input  *rpc_v0.AccountsParam `json:"input"`
		Output *core_types.AccountList   `json:"output"`
	}

	GetStorageData struct {
		Input  *rpc_v0.AddressParam `json:"input"`
		Output *core_types.Storage      `json:"output"`
	}

	GetStorageAtData struct {
		Input  *rpc_v0.StorageAtParam `json:"input"`
		Output *core_types.StorageItem    `json:"output"`
	}

	GenPrivAccountData struct {
		Output *account.PrivAccount `json:"output"`
	}

	GetBlockchainInfoData struct {
		Output *core_types.BlockchainInfo `json:"output"`
	}

	GetChainIdData struct {
		Output *core_types.ChainId `json:"output"`
	}

	GetGenesisHashData struct {
		Output *core_types.GenesisHash `json:"output"`
	}

	GetLatestBlockHeightData struct {
		Output *core_types.LatestBlockHeight `json:"output"`
	}

	GetLatestBlockData struct {
		Output *mintTypes.Block `json:"output"`
	}

	GetBlockData struct {
		Input  *rpc_v0.HeightParam `json:"input"`
		Output *mintTypes.Block  `json:"output"`
	}

	GetBlocksData struct {
		Input  *rpc_v0.BlocksParam  `json:"input"`
		Output *core_types.Blocks `json:"output"`
	}

	GetConsensusStateData struct {
		Output *core_types.ConsensusState `json:"output"`
	}

	GetValidatorsData struct {
		Output *core_types.ValidatorList `json:"output"`
	}

	GetNetworkInfoData struct {
		Output *core_types.NetworkInfo `json:"output"`
	}

	GetClientVersionData struct {
		Output *core_types.ClientVersion `json:"output"`
	}

	GetMonikerData struct {
		Output *core_types.Moniker `json:"output"`
	}

	IsListeningData struct {
		Output *core_types.Listening `json:"output"`
	}

	GetListenersData struct {
		Output *core_types.Listeners `json:"output"`
	}

	GetPeersData struct {
		Output []*core_types.Peer `json:"output"`
	}

	GetPeerData struct {
		Input  *rpc_v0.PeerParam  `json:"input"`
		Output *core_types.Peer `json:"output"`
	}

	TransactData struct {
		Input  *rpc_v0.TransactParam `json:"input"`
		Output *core_types.Receipt `json:"output"`
	}

	TransactCreateData struct {
		Input  *rpc_v0.TransactParam `json:"input"`
		Output *core_types.Receipt `json:"output"`
	}

	GetUnconfirmedTxsData struct {
		Output *core_types.UnconfirmedTxs `json:"output"`
	}

	CallCodeData struct {
		Input  *rpc_v0.CallCodeParam `json:"input"`
		Output *core_types.Call    `json:"output"`
	}

	CallData struct {
		Input  *rpc_v0.CallParam  `json:"input"`
		Output *core_types.Call `json:"output"`
	}

	EventSubscribeData struct {
		Input  *rpc_v0.EventIdParam   `json:"input"`
		Output *event.EventSub `json:"output"`
	}

	EventUnsubscribeData struct {
		Input  *rpc_v0.SubIdParam       `json:"input"`
		Output *event.EventUnsub `json:"output"`
	}

	TransactNameRegData struct {
		Input  *rpc_v0.TransactNameRegParam `json:"input"`
		Output *core_types.Receipt        `json:"output"`
	}

	GetNameRegEntryData struct {
		Input  *rpc_v0.NameRegEntryParam `json:"input"`
		Output *types.NameRegEntry     `json:"output"`
	}

	GetNameRegEntriesData struct {
		Input  *rpc_v0.FilterListParam       `json:"input"`
		Output *core_types.ResultListNames `json:"output"`
	}

	/*
		EventPollData struct {
			Input  *rpc_v0.SubIdParam  `json:"input"`
			Output *event.PollResponse `json:"output"`
		}
	*/

	TestData struct {
		ChainData            *ChainData `json:"chain_data"`
		GetAccount           *GetAccountData
		GetAccounts          *GetAccountsData
		GetStorage           *GetStorageData
		GetStorageAt         *GetStorageAtData
		GenPrivAccount       *GenPrivAccountData
		GetBlockchainInfo    *GetBlockchainInfoData
		GetChainId           *GetChainIdData
		GetGenesisHash       *GetGenesisHashData
		GetLatestBlockHeight *GetLatestBlockHeightData
		GetLatestBlock       *GetLatestBlockData
		GetBlock             *GetBlockData
		GetBlocks            *GetBlocksData
		GetConsensusState    *GetConsensusStateData
		GetValidators        *GetValidatorsData
		GetNetworkInfo       *GetNetworkInfoData
		GetClientVersion     *GetClientVersionData
		GetMoniker           *GetMonikerData
		IsListening          *IsListeningData
		GetListeners         *GetListenersData
		GetPeers             *GetPeersData
		Transact             *TransactData
		TransactCreate       *TransactCreateData
		TransactNameReg      *TransactNameRegData
		GetUnconfirmedTxs    *GetUnconfirmedTxsData
		CallCode             *CallCodeData
		Call                 *CallData
		EventSubscribe       *EventSubscribeData
		EventUnsubscribe     *EventUnsubscribeData
		GetNameRegEntry      *GetNameRegEntryData
		GetNameRegEntries    *GetNameRegEntriesData
		// GetPeer              *GetPeerData
		// EventPoll            *EventPollData
	}
)

func LoadTestData() *TestData {
	codec := rpc_v0.NewTCodec()
	testData := &TestData{}
	if err := codec.DecodeBytes(testData, []byte(testDataJson)); err != nil {
		panic(err)
	}
	return testData
}
