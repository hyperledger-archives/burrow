package test

import (
	"github.com/tendermint/tendermint/state"
	edb "github.com/eris-ltd/erisdb/erisdb"
	ess "github.com/eris-ltd/erisdb/erisdb/erisdbss"
)


var privValidator = []byte(`{
    "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
    "pub_key": [1, "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"],
    "priv_key": [1, "6B72D45EB65F619F11CE580C8CAED9E0BADC774E9C9C334687A65DCBAD2C4151CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"],
    "last_height": 0,
    "last_round": 0,
    "last_step": 0
}`);

var genesisDoc = []byte(`{
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
}`);

var serverDuration uint = 100

type TestData struct {
	requestData *ess.RequestData
}

func LoadTestData() *TestData {
	codec := edb.NewTCodec()
	pvd := &state.PrivValidator{}
	_ = codec.DecodeBytes(pvd, privValidator)
	genesis := &state.GenesisDoc{}
	_ = codec.DecodeBytes(genesis, genesisDoc)
	td := &TestData{}
	td.requestData = &ess.RequestData{pvd, genesis, serverDuration}
	return td
}