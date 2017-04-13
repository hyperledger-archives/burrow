package config

import (
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var complexConfig *LoggingConfig = &LoggingConfig{
	InfoSink: Sink().
		SetOutput(StderrOutput()),
	InfoAndTraceSink: Sink().
		SetTransform(LabelTransform(false, "Info", "Trace")).
		AddSinks(
			Sink().
				SetOutput(StdoutOutput()).
				SetTransform(FilterTransform(ExcludeWhenAnyMatches,
					"Foo", "Bars")).
				AddSinks(
					Sink().
						SetOutput(RemoteSyslogOutput("Eris-db", "tcp://example.com:6514")),
					Sink().
						SetOutput(StdoutOutput()),
				),
		),
}

func TestLoggingConfig_String(t *testing.T) {
	lc := new(LoggingConfig)
	toml.Decode(complexConfig.TOMLString(), lc)
	assert.Equal(t, complexConfig, lc)
}

func TestReadViperConfig(t *testing.T) {
	conf := viper.New()
	conf.SetConfigType("toml")
	conf.ReadConfig(strings.NewReader(complexConfig.TOMLString()))
	lc, err := LoggingConfigFromMap(conf.AllSettings())
	assert.NoError(t, err)
	assert.Equal(t, complexConfig, lc)
}
