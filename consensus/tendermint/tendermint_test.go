package tendermint

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
	"github.com/tendermint/tendermint/p2p"
)

func TestMarshalNodeKey(t *testing.T) {
	cdc := amino.NewCodec()
	cryptoAmino.RegisterAmino(cdc)
	nodeKey := NewNodeKey()

	bsAmino, err := cdc.MarshalJSON(nodeKey)
	require.NoError(t, err)

	file, err := ioutil.TempFile(os.TempDir(), "nodeKey-")
	require.NoError(t, err)
	defer os.Remove(file.Name())
	_, err = file.Write(bsAmino)
	require.NoError(t, err)
	nk, err := p2p.LoadNodeKey(file.Name())
	require.NoError(t, err)
	require.Equal(t, *nodeKey, *nk)

	bs, err := json.Marshal(nodeKey)
	require.NoError(t, err)
	nk = new(p2p.NodeKey)
	nk.PrivKey = new(ed25519.PrivKeyEd25519)
	err = json.Unmarshal(bs, nk)
	require.NoError(t, err)
	require.Equal(t, nodeKey.PrivKey, *nk.PrivKey.(*ed25519.PrivKeyEd25519))
}
