package presets

import (
	"testing"

	"strconv"

	"github.com/hyperledger/burrow/logging/logconfig"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSinkConfig(t *testing.T) {
	builtSink, err := BuildSinkConfig(IncludeAny, Info, Stdout, Terminal, Down, Down, Info, Stdout, Up, Info, Stderr)
	require.NoError(t, err)
	expectedSink := logconfig.Sink().
		SetTransform(logconfig.FilterTransform(logconfig.IncludeWhenAnyMatches,
			structure.ChannelKey, structure.InfoChannelName)).SetOutput(logconfig.StdoutOutput().SetFormat(loggers.TerminalFormat)).AddSinks(
		logconfig.Sink().SetTransform(logconfig.FilterTransform(logconfig.NoFilterMode,
			structure.ChannelKey, structure.InfoChannelName)).SetOutput(logconfig.StderrOutput()).AddSinks(
			logconfig.Sink().SetTransform(logconfig.FilterTransform(logconfig.NoFilterMode,
				structure.ChannelKey, structure.InfoChannelName)).SetOutput(logconfig.StdoutOutput())))

	//fmt.Println(config.JSONString(expectedSink), "\n", config.JSONString(builtSink))
	assert.Equal(t, logconfig.JSONString(expectedSink), logconfig.JSONString(builtSink))
}

func TestMinimalPreset(t *testing.T) {
	builtSink, err := BuildSinkConfig(Minimal, Stderr)
	require.NoError(t, err)
	expectedSink := logconfig.Sink().
		AddSinks(logconfig.Sink().SetTransform(logconfig.PruneTransform(structure.TraceKey, structure.RunId)).
			AddSinks(logconfig.Sink().SetTransform(logconfig.FilterTransform(logconfig.IncludeWhenAllMatch,
				structure.ChannelKey, structure.InfoChannelName)).
				AddSinks(logconfig.Sink().SetTransform(logconfig.FilterTransform(logconfig.ExcludeWhenAnyMatches,
					structure.ComponentKey, "Tendermint",
					"module", "p2p",
					"module", "mempool")).SetOutput(logconfig.StderrOutput()))))
	//fmt.Println(config.TOMLString(expectedSink), "\n", config.TOMLString(builtSink))
	assert.Equal(t, logconfig.TOMLString(expectedSink), logconfig.TOMLString(builtSink))
}

func TestFileOutput(t *testing.T) {
	path := "foo.log"
	builtSink, err := BuildSinkConfig(Down, File, path, JSON)
	require.NoError(t, err)
	expectedSink := logconfig.Sink().
		AddSinks(logconfig.Sink().SetOutput(logconfig.FileOutput(path).SetFormat(loggers.JSONFormat)))
	//fmt.Println(config.TOMLString(expectedSink), "\n", config.TOMLString(builtSink))
	assert.Equal(t, logconfig.TOMLString(expectedSink), logconfig.TOMLString(builtSink))
}

func TestCaptureLoggerNormalLogger(t *testing.T) {
	path := "/dev/termination-log"
	name := "hello"
	buffer := 10000
	builtSink, err := BuildSinkConfig(Capture, name, strconv.Itoa(buffer), File, path, JSON, Top, IncludeAny, Info, Stderr, JSON)
	require.NoError(t, err)
	expectedSink := logconfig.Sink().
		AddSinks(
			logconfig.Sink().SetTransform(logconfig.CaptureTransform(name, buffer, false)).
				SetOutput(logconfig.FileOutput(path).SetFormat(loggers.JSONFormat)),
			logconfig.Sink().SetTransform(logconfig.FilterTransform(logconfig.IncludeWhenAnyMatches,
				structure.ChannelKey, structure.InfoChannelName)).SetOutput(logconfig.StderrOutput().
				SetFormat(loggers.JSONFormat)))
	assert.Equal(t, logconfig.TOMLString(expectedSink), logconfig.TOMLString(builtSink))
}
