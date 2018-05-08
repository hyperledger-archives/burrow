package config

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingConfig_String(t *testing.T) {
	complexConfig := &LoggingConfig{
		RootSink: Sink().
			SetTransform(LabelTransform(false, "Info", "Trace")).
			AddSinks(
				Sink().
					SetOutput(StdoutOutput()).
					SetTransform(FilterTransform(ExcludeWhenAnyMatches,
						"Foo", "Bars")).
					AddSinks(
						Sink().
							SetOutput(StderrOutput()),
						Sink().
							SetOutput(StdoutOutput()),
					),
			),
	}
	lc := new(LoggingConfig)
	_, err := toml.Decode(complexConfig.TOMLString(), lc)
	require.NoError(t, err)
	assert.Equal(t, complexConfig, lc)
}
