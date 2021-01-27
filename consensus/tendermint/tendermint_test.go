package tendermint

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/p2p"
)

func TestMarshalNodeKey(t *testing.T) {
	nodeKey := NewNodeKey()

	tmbs, err := tmjson.Marshal(nodeKey)
	require.NoError(t, err)

	file, err := ioutil.TempFile(os.TempDir(), "nodeKey-")
	require.NoError(t, err)
	defer os.Remove(file.Name())
	_, err = file.Write(tmbs)
	require.NoError(t, err)
	nk, err := p2p.LoadNodeKey(file.Name())
	require.NoError(t, err)
	require.Equal(t, *nodeKey, *nk)

	bs, err := json.Marshal(nodeKey)
	require.NoError(t, err)
	nk = new(p2p.NodeKey)
	nk.PrivKey = new(ed25519.PrivKey)
	err = json.Unmarshal(bs, nk)
	require.NoError(t, err)
	require.Equal(t, nodeKey.PrivKey, *nk.PrivKey.(*ed25519.PrivKey))
}
