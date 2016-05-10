package rpctest

// priv keys generated deterministically eg rpc/tests/helpers.go
var defaultGenesis = `{
  "chain_id" : "tendermint_test",
  "accounts": [
    {
	    "address": "E9B5D87313356465FAE33C406CE2C2979DE60BCB",
	    "amount": 200000000
    },
    {
	    "address": "DFE4AFFA4CEE17CD01CB9E061D77C3ECED29BD88",
	    "amount": 200000000
    },
    {
	    "address": "F60D30722E7B497FA532FB3207C3FB29C31B1992",
	    "amount": 200000000
    },
    {
	    "address": "336CB40A5EB92E496E19B74FDFF2BA017C877FD6",
	    "amount": 200000000
    },
    {
	    "address": "D218F0F439BF0384F6F5EF8D0F8B398D941BD1DC",
	    "amount": 200000000
    }
  ],
  "validators": [
    {
      "pub_key": [1, "583779C3BFA3F6C7E23C7D830A9C3D023A216B55079AD38BFED1207B94A19548"],
      "amount": 1000000,
      "unbond_to": [
        {
          "address": "E9B5D87313356465FAE33C406CE2C2979DE60BCB",
          "amount":  100000
        }
      ]
    }
  ]
}`

var defaultPrivValidator = `{
  "address": "1D7A91CB32F758A02EBB9BE1FB6F8DEE56F90D42",
	"pub_key": [1,"06FBAC4E285285D1D91FCBC7E91C780ADA11516F67462340B3980CE2B94940E8"],
	"priv_key": [1,"C453604BD6480D5538B4C6FD2E3E314B5BCE518D75ADE4DA3DA85AB8ADFD819606FBAC4E285285D1D91FCBC7E91C780ADA11516F67462340B3980CE2B94940E8"],
	"last_height":0,
	"last_round":0,
	"last_step":0
}`
