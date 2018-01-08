package tm

import (
	"testing"

	"github.com/hyperledger/burrow/rpc"
	"github.com/stretchr/testify/require"
)

func testChainId(chainName string) (*rpc.ResultChainId, error) {
	return &rpc.ResultChainId{
		ChainName:   chainName,
		ChainId:     "Foos",
		GenesisHash: []byte{},
	}, nil
}

func TestWrapFuncBurrowResult(t *testing.T) {
	f, err := wrapFuncBurrowResult(testChainId)
	require.NoError(t, err)
	fOut, ok := f.(func(string) (rpc.Result, error))
	require.True(t, ok, "must be able to cast to function type")
	br, err := fOut("Blum")
	require.NoError(t, err)
	bs, err := br.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, `{"type":"result_chain_id","data":{"chain_name":"Blum","chain_id":"Foos","genesis_hash":""}}`,
		string(bs))
}
