package config

import (
	"fmt"
	"os"

	"net/url"

	"github.com/eapache/channels"
	kitlog "github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/hyperledger/burrow/logging/structure"
)

// This file contains definitions for a configurable output graph for the
// logging system.

type source string
type outputType string
type transformType string
type filterMode string

const (
	// OutputType
	NoOutput outputType = ""
	Graylog  outputType = "graylog"
	Syslog   outputType = "syslog"
	File     outputType = "file"
	Stdout   outputType = "stdout"
	Stderr   outputType = "stderr"

	// TransformType
	NoTransform transformType = ""
	// Filter log lines
	Filter transformType = "filter"
	// Remove key-val pairs from each log line
	Prune transformType = "prune"
	// Add key value pairs to each log line
	Label   transformType = "label"
	Capture transformType = "capture"
	// TODO [Silas]: add 'flush on exit' transform which flushes the buffer of
	// CaptureLogger to its OutputLogger a non-passthrough capture when an exit
	// signal is detected or some other exceptional thing happens

	IncludeWhenAllMatch   filterMode = "include_when_all_match"
	IncludeWhenAnyMatches filterMode = "include_when_any_matches"
	ExcludeWhenAllMatch   filterMode = "exclude_when_all_match"
	ExcludeWhenAnyMatches filterMode = "exclude_when_any_matches"
)

// Only include log lines matching the filter so negate the predicate in filter
func (mode filterMode) Include() bool {
	switch mode {
	case IncludeWhenAllMatch, IncludeWhenAnyMatches:
		return true
	default:
		return false
	}
}

// The predicate should only evaluate true if all the key value predicates match
func (mode filterMode) MatchAll() bool {
	switch mode {
	case IncludeWhenAllMatch, ExcludeWhenAllMatch:
		return true
	default:
		return false
	}
}

// Exclude log lines that match the predicate
func (mode filterMode) Exclude() bool {
	return !mode.Include()
}

// The predicate should evaluate true if at least one of the key value predicates matches
func (mode filterMode) MatchAny() bool {
	return !mode.MatchAny()
}

// Sink configuration types
type (
	// Outputs
	GraylogConfig struct {
	}

	SyslogConfig struct {
		Url string `toml:"url"`
		Tag string `toml:"tag"`
	}

	FileConfig struct {
		Path string `toml:"path"`
	}

	OutputConfig struct {
		OutputType outputType `toml:"output_type"`
		Format     string     `toml:"format,omitempty"`
		*GraylogConfig
		*FileConfig
		*SyslogConfig
	}

	// Transforms
	LabelConfig struct {
		Labels map[string]string `toml:"labels"`
		Prefix bool              `toml:"prefix"`
	}

	PruneConfig struct {
		Keys []string `toml:"keys"`
	}

	CaptureConfig struct {
		Name        string `toml:"name"`
		BufferCap   int    `toml:"buffer_cap"`
		Passthrough bool   `toml:"passthrough"`
	}

	// Generates true if KeyRegex matches a log line key and ValueRegex matches that key's value.
	// If ValueRegex is empty then returns true if any key matches
	// If KeyRegex is empty then returns true if any value matches
	KeyValuePredicateConfig struct {
		KeyRegex   string `toml:"key_regex"`
		ValueRegex string `toml:"value_regex"`
	}

	// Filter types
	FilterConfig struct {
		FilterMode filterMode `toml:"filter_mode"`
		// Predicates to match a log line against using FilterMode
		Predicates []*KeyValuePredicateConfig `toml:"predicates"`
	}

	TransformConfig struct {
		TransformType transformType `toml:"transform_type"`
		*LabelConfig
		*PruneConfig
		*CaptureConfig
		*FilterConfig
	}

	// Sink
	// A Sink describes a logger that logs to zero or one output and logs to zero or more child sinks.
	// before transmitting its log it applies zero or one transforms to the stream of log lines.
	// by chaining together many Sinks arbitrary transforms to and multi
	SinkConfig struct {
		Transform *TransformConfig `toml:"transform"`
		Sinks     []*SinkConfig    `toml:"sinks"`
		Output    *OutputConfig    `toml:"output"`
	}
)

// Builders
func Sink() *SinkConfig {
	return &SinkConfig{}
}

func (sinkConfig *SinkConfig) AddSinks(sinks ...*SinkConfig) *SinkConfig {
	sinkConfig.Sinks = append(sinkConfig.Sinks, sinks...)
	return sinkConfig
}

func (sinkConfig *SinkConfig) SetTransform(transform *TransformConfig) *SinkConfig {
	sinkConfig.Transform = transform
	return sinkConfig
}

func (sinkConfig *SinkConfig) SetOutput(output *OutputConfig) *SinkConfig {
	sinkConfig.Output = output
	return sinkConfig
}

func StdoutOutput() *OutputConfig {
	return &OutputConfig{
		OutputType: Stdout,
	}
}

func StderrOutput() *OutputConfig {
	return &OutputConfig{
		OutputType: Stderr,
	}
}

func SyslogOutput(tag string) *OutputConfig {
	return RemoteSyslogOutput(tag, "")
}

func FileOutput(path string) *OutputConfig {
	return &OutputConfig{
		OutputType: File,
		FileConfig: &FileConfig{
			Path: path,
		},
	}
}

func RemoteSyslogOutput(tag, remoteUrl string) *OutputConfig {
	return &OutputConfig{
		OutputType: Syslog,
		SyslogConfig: &SyslogConfig{
			Url: remoteUrl,
			Tag: tag,
		},
	}
}

func CaptureTransform(name string, bufferCap int, passthrough bool) *TransformConfig {
	return &TransformConfig{
		TransformType: Capture,
		CaptureConfig: &CaptureConfig{
			Name:        name,
			BufferCap:   bufferCap,
			Passthrough: passthrough,
		},
	}
}

func LabelTransform(prefix bool, labelKeyvals ...string) *TransformConfig {
	length := len(labelKeyvals) / 2
	labels := make(map[string]string, length)
	for i := 0; i < 2*length; i += 2 {
		labels[labelKeyvals[i]] = labelKeyvals[i+1]
	}
	return &TransformConfig{
		TransformType: Label,
		LabelConfig: &LabelConfig{
			Prefix: prefix,
			Labels: labels,
		},
	}
}

func PruneTransform(keys ...string) *TransformConfig {
	return &TransformConfig{
		TransformType: Prune,
		PruneConfig: &PruneConfig{
			Keys: keys,
		},
	}
}

func FilterTransform(fmode filterMode, keyvalueRegexes ...string) *TransformConfig {
	length := len(keyvalueRegexes) / 2
	predicates := make([]*KeyValuePredicateConfig, length)
	for i := 0; i < length; i++ {
		kv := i * 2
		predicates[i] = &KeyValuePredicateConfig{
			KeyRegex:   keyvalueRegexes[kv],
			ValueRegex: keyvalueRegexes[kv+1],
		}
	}
	return &TransformConfig{
		TransformType: Filter,
		FilterConfig: &FilterConfig{
			FilterMode: fmode,
			Predicates: predicates,
		},
	}
}

// Logger formation
func (sinkConfig *SinkConfig) BuildLogger() (kitlog.Logger, map[string]*loggers.CaptureLogger, error) {
	return BuildLoggerFromSinkConfig(sinkConfig, make(map[string]*loggers.CaptureLogger))
}

func BuildLoggerFromSinkConfig(sinkConfig *SinkConfig,
	captures map[string]*loggers.CaptureLogger) (kitlog.Logger, map[string]*loggers.CaptureLogger, error) {
	if sinkConfig == nil {
		return kitlog.NewNopLogger(), captures, nil
	}
	numSinks := len(sinkConfig.Sinks)
	outputLoggers := make([]kitlog.Logger, numSinks, numSinks+1)
	// We need a depth-first post-order over the output loggers so we'll keep
	// recurring into children sinks we reach a terminal sink (with no children)
	for i, sc := range sinkConfig.Sinks {
		l, captures, err := BuildLoggerFromSinkConfig(sc, captures)
		if err != nil {
			return nil, captures, err
		}
		outputLoggers[i] = l
	}

	// Grab the outputs after we have terminated any children sinks above
	if sinkConfig.Output != nil && sinkConfig.Output.OutputType != NoOutput {
		l, err := BuildOutputLogger(sinkConfig.Output)
		if err != nil {
			return nil, captures, err
		}
		outputLoggers = append(outputLoggers, l)
	}

	outputLogger := loggers.NewMultipleOutputLogger(outputLoggers...)

	if sinkConfig.Transform != nil && sinkConfig.Transform.TransformType != NoTransform {
		return BuildTransformLogger(sinkConfig.Transform, captures, outputLogger)
	}
	return outputLogger, captures, nil
}

func BuildOutputLogger(outputConfig *OutputConfig) (kitlog.Logger, error) {
	switch outputConfig.OutputType {
	case NoOutput:
		return kitlog.NewNopLogger(), nil
	//case Graylog:
	case Syslog:
		urlString := outputConfig.SyslogConfig.Url
		if urlString != "" {
			remoteUrl, err := url.Parse(urlString)
			if err != nil {
				return nil, fmt.Errorf("Error parsing remote syslog URL: %s, "+
					"error: %s",
					urlString, err)
			}
			return loggers.NewRemoteSyslogLogger(remoteUrl,
				outputConfig.SyslogConfig.Tag, outputConfig.Format)
		}
		return loggers.NewSyslogLogger(outputConfig.SyslogConfig.Tag,
			outputConfig.Format)
	case Stdout:
		return loggers.NewStreamLogger(os.Stdout, outputConfig.Format), nil
	case Stderr:
		return loggers.NewStreamLogger(os.Stderr, outputConfig.Format), nil
	case File:
		return loggers.NewFileLogger(outputConfig.FileConfig.Path, outputConfig.Format)
	default:
		return nil, fmt.Errorf("Could not build logger for output: '%s'",
			outputConfig.OutputType)
	}
}

func BuildTransformLogger(transformConfig *TransformConfig,
	captures map[string]*loggers.CaptureLogger,
	outputLogger kitlog.Logger) (kitlog.Logger, map[string]*loggers.CaptureLogger, error) {
	switch transformConfig.TransformType {
	case NoTransform:
		return outputLogger, captures, nil
	case Label:
		keyvals := make([]interface{}, 0, len(transformConfig.Labels)*2)
		for k, v := range transformConfig.LabelConfig.Labels {
			keyvals = append(keyvals, k, v)
		}
		if transformConfig.LabelConfig.Prefix {
			return kitlog.WithPrefix(outputLogger, keyvals...), captures, nil
		} else {
			return kitlog.With(outputLogger, keyvals...), captures, nil
		}
	case Prune:
		keys := make([]interface{}, len(transformConfig.PruneConfig.Keys))
		for i, k := range transformConfig.PruneConfig.Keys {
			keys[i] = k
		}
		return kitlog.LoggerFunc(func(keyvals ...interface{}) error {
			return outputLogger.Log(structure.RemoveKeys(keyvals, keys...)...)
		}), captures, nil

	case Capture:
		name := transformConfig.CaptureConfig.Name
		if _, ok := captures[name]; ok {
			return nil, captures, fmt.Errorf("Could not register new logging capture since name '%s' already "+
				"registered", name)
		}
		// Create a capture logger according to configuration (it may tee the output)
		// or capture it to be flushed later
		captureLogger := loggers.NewCaptureLogger(outputLogger,
			channels.BufferCap(transformConfig.CaptureConfig.BufferCap),
			transformConfig.CaptureConfig.Passthrough)
		// Register the capture
		captures[name] = captureLogger
		// Pass it upstream to be logged to
		return captureLogger, captures, nil
	case Filter:
		predicate, err := BuildFilterPredicate(transformConfig.FilterConfig)
		if err != nil {
			return nil, captures, fmt.Errorf("Could not build filter predicate: '%s'", err)
		}
		return loggers.NewFilterLogger(outputLogger, predicate), captures, nil
	default:
		return nil, captures, fmt.Errorf("Could not build logger for transform: '%s'", transformConfig.TransformType)
	}
}
