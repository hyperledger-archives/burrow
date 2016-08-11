package test

// priv keys generated deterministically eg rpc/tests/shared.go
var defaultGenesis = `{
  "chain_id" : "MyChainId",
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
