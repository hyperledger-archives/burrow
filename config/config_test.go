package config

import (
	"testing"

	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/genesis"
	"github.com/stretchr/testify/require"
)

func TestBurrowConfigSerialise(t *testing.T) {
	conf := &BurrowConfig{
		GenesisDoc: &genesis.GenesisDoc{
			ChainName: "Foo",
		},
	}
	confOut := new(BurrowConfig)
	jsonString := conf.JSONString()
	err := source.FromJSONString(jsonString, confOut)
	require.NoError(t, err)
	require.Equal(t, jsonString, confOut.JSONString())
}
