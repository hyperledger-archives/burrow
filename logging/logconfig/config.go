package logconfig

import (
	"bytes"
	"fmt"

	"github.com/eapache/channels"
	"github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"

	"encoding/json"

	"github.com/BurntSushi/toml"
	"github.com/hyperledger/burrow/logging/loggers"
)

type LoggingConfig struct {
	RootSink *SinkConfig `toml:",omitempty"`
	// Trace debug is very noisy - mostly from Tendermint
	Trace bool
	// Send to a channel - will not affect progress if logging graph is intensive but output will lag and some logs
	// may be missed in shutdown
	NonBlocking bool
}

// For encoding a top-level '[logging]' TOML table
type LoggingConfigWrapper struct {
	Logging *LoggingConfig `toml:",omitempty"`
}

func DefaultNodeLoggingConfig() *LoggingConfig {
	// Output only Burrow messages on stdout
	return &LoggingConfig{
		RootSink: Sink().
			SetTransform(FilterTransform(ExcludeWhenAnyMatches, structure.ComponentKey, structure.Tendermint)).
			SetOutput(StdoutOutput().SetFormat(loggers.JSONFormat)),
	}
}

// Provide a defeault logging config
func New() *LoggingConfig {
	return &LoggingConfig{
		NonBlocking: false,
		RootSink:    Sink().SetOutput(StderrOutput().SetFormat(JSONFormat)),
	}
}

func (lc *LoggingConfig) Root(configure func(sink *SinkConfig) *SinkConfig) *LoggingConfig {
	lc.RootSink = configure(Sink())
	return lc
}

// Returns the TOML for a top-level logging config wrapped with [logging]
func (lc *LoggingConfig) RootTOMLString() string {
	return TOMLString(LoggingConfigWrapper{lc})
}

func (lc *LoggingConfig) TOMLString() string {
	return TOMLString(lc)
}

func (lc *LoggingConfig) RootJSONString() string {
	return JSONString(LoggingConfigWrapper{lc})
}

func (lc *LoggingConfig) JSONString() string {
	return JSONString(lc)
}

// Obtain a logger from this LoggingConfig
func (lc *LoggingConfig) NewLogger() (*logging.Logger, error) {
	outputLogger, errCh, err := newLogger(lc)
	if err != nil {
		return nil, err
	}
	logger := logging.NewLogger(outputLogger)
	if !lc.Trace {
		logger.Trace = log.NewNopLogger()
	}
	go func() {
		err := <-errCh.Out()
		if err != nil {
			fmt.Printf("Logging error: %v", err)
		}
	}()
	return logger, nil
}

// Hot swap logging config by replacing output loggers built from this LoggingConfig
func (lc *LoggingConfig) UpdateLogger(logger *logging.Logger) (channels.Channel, error) {
	outputLogger, errCh, err := newLogger(lc)
	if err != nil {
		return channels.NewDeadChannel(), err
	}
	logger.SwapOutput(outputLogger)
	return errCh, nil
}

// Helpers
func newLogger(loggingConfig *LoggingConfig) (log.Logger, channels.Channel, error) {
	outputLogger, _, err := loggingConfig.RootSink.BuildLogger()
	if err != nil {
		return nil, nil, err
	}
	var errCh channels.Channel = channels.NewDeadChannel()
	var logger log.Logger = loggers.BurrowFormatLogger(outputLogger)
	if loggingConfig.NonBlocking {
		logger, errCh = loggers.NonBlockingLogger(logger)
		return logger, errCh, nil
	}
	return logger, errCh, err
}

func TOMLString(v interface{}) string {
	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(buf)
	err := encoder.Encode(v)
	if err != nil {
		// Seems like a reasonable compromise to make the string function clean
		return fmt.Sprintf("Error encoding TOML: %s", err)
	}
	return buf.String()
}

func JSONString(v interface{}) string {
	bs, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return fmt.Sprintf("Error encoding JSON: %s", err)
	}
	return string(bs)
}
