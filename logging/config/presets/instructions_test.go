package presets

import (
	"testing"

	"fmt"

	"github.com/hyperledger/burrow/logging/config"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/logging/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSinkConfig(t *testing.T) {
	builtSink, err := BuildSinkConfig(IncludeAny, Info, Stdout, Terminal, Down, Down, Info, Stdout, Up, Info, Stderr)
	require.NoError(t, err)
	expectedSink := config.Sink().
		SetTransform(config.FilterTransform(config.IncludeWhenAnyMatches,
			structure.ChannelKey, types.InfoChannelName)).SetOutput(config.StdoutOutput().SetFormat(loggers.TerminalFormat)).AddSinks(
		config.Sink().SetTransform(config.FilterTransform(config.NoFilterMode,
			structure.ChannelKey, types.InfoChannelName)).SetOutput(config.StderrOutput()).AddSinks(
			config.Sink().SetTransform(config.FilterTransform(config.NoFilterMode,
				structure.ChannelKey, types.InfoChannelName)).SetOutput(config.StdoutOutput())))

	//fmt.Println(config.JSONString(expectedSink), "\n", config.JSONString(builtSink))
	assert.Equal(t, config.JSONString(expectedSink), config.JSONString(builtSink))
}

func TestMinimalPreset(t *testing.T) {
	builtSink, err := BuildSinkConfig(Minimal, Stderr)
	require.NoError(t, err)
	expectedSink := config.Sink().
		AddSinks(config.Sink().SetTransform(config.PruneTransform(structure.TraceKey, structure.RunId)).
			AddSinks(config.Sink().SetTransform(config.FilterTransform(config.IncludeWhenAllMatch,
				structure.ChannelKey, types.InfoChannelName)).
				AddSinks(config.Sink().SetTransform(config.FilterTransform(config.ExcludeWhenAnyMatches,
					structure.ComponentKey, "Tendermint",
					"module", "p2p",
					"module", "mempool")).SetOutput(config.StderrOutput()))))
	//fmt.Println(config.TOMLString(expectedSink), "\n", config.TOMLString(builtSink))
	assert.Equal(t, config.TOMLString(expectedSink), config.TOMLString(builtSink))
}
