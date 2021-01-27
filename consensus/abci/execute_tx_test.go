package abci

import (
	"bytes"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/abci/types"
)

func TestWithEvents(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.NewLogger(log.NewLogfmtLogger(&buf))
	kvp := types.EventAttribute{Key: []byte("foo"), Value: []byte("bar")}
	event := types.Event{Type: "event", Attributes: []types.EventAttribute{kvp}}
	events := []types.Event{event}
	logger = WithEvents(logger, events)
	logger.InfoMsg("hello, world")
	require.Equal(t, "log_channel=Info event=foo:bar message=\"hello, world\"\n", buf.String())
}
