package tendermint

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-amino"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
)

func TestMarshalNodeKey(t *testing.T) {

	cdc := amino.NewCodec()
	cryptoAmino.RegisterAmino(cdc)

	nodeKey := NewNodeKey()

	bsAmino, err := cdc.MarshalJSON(nodeKey)
	require.NoError(t, err)
	fmt.Println(string(bsAmino))

	bs, err := json.Marshal(nodeKey)
	require.NoError(t, err)

	fmt.Println(string(bs))
}
