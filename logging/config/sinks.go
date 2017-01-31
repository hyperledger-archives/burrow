package config

import (
	"fmt"
	"os"

	"github.com/eapache/channels"
	"github.com/eris-ltd/eris-db/logging/loggers"
	kitlog "github.com/go-kit/kit/log"
)

type source string
type outputType string
type transformType string
type filterMode string

const (
	// OutputType
	NoOutput outputType = ""
	Graylog  outputType = "Graylog"
	Syslog   outputType = "Syslog"
	File     outputType = "File"
	Stdout   outputType = "Stdout"
	Stderr   outputType = "Stderr"

	// TransformType
	NoTransform transformType = ""
	// Filter log lines
	Filter transformType = "Filter"
	// Remove key-val pairs from each log line
	Prune   transformType = "Prune"
	Capture transformType = "Capture"
	Label   transformType = "Label"

	IncludeWhenAllMatch   filterMode = "IncludeWhenAllMatch"
	IncludeWhenAnyMatches filterMode = "IncludeWhenAnyMatches"
	ExcludeWhenAllMatch   filterMode = "ExcludeWhenAllMatch"
	ExcludeWhenAnyMatches filterMode = "ExcludeWhenAnyMatches"
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
	}

	FileConfig struct {
		Path string
	}

	OutputConfig struct {
		OutputType outputType
		*GraylogConfig
		*FileConfig
		*SyslogConfig
	}

	// Transforms
	LabelConfig struct {
		Labels map[string]string
		Prefix bool
	}

	CaptureConfig struct {
		Name        string
		BufferCap   int
		Passthrough bool
	}

	// Generates true if KeyRegex matches a log line key and ValueRegex matches that key's value.
	// If ValueRegex is empty then returns true if any key matches
	// If KeyRegex is empty then returns true if any value matches
	KeyValuePredicateConfig struct {
		KeyRegex   string
		ValueRegex string
	}

	// Filter types

	FilterConfig struct {
		FilterMode filterMode
		// Predicates to match a log line against using FilterMode
		Predicates []*KeyValuePredicateConfig
	}

	TransformConfig struct {
		TransformType transformType
		*LabelConfig
		*CaptureConfig
		*FilterConfig
	}

	// Sink
	// A Sink describes a logger that logs to zero or one output and logs to zero or more child sinks.
	// before transmitting its log it applies zero or one transforms to the stream of log lines.
	// by chaining together many Sinks arbitrary transforms to and multi
	SinkConfig struct {
		Transform *TransformConfig
		Sinks     []*SinkConfig
		Output    *OutputConfig
	}

	LoggingConfig struct {
		InfoSink         *SinkConfig
		InfoAndTraceSink *SinkConfig
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

func CaptureTransform(name string, bufferCap int, passthrough bool) *TransformConfig {
	return &TransformConfig{
		TransformType: Capture,
		CaptureConfig: &CaptureConfig{
			Name:      name,
			BufferCap: bufferCap,
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

func (sinkConfig *SinkConfig) BuildLogger() (kitlog.Logger, map[string]*loggers.CaptureLogger, error) {
	return BuildLoggerFromSinkConfig(sinkConfig, make(map[string]*loggers.CaptureLogger))
}

// Logger formation

func BuildLoggerFromSinkConfig(sinkConfig *SinkConfig,
	captures map[string]*loggers.CaptureLogger) (kitlog.Logger, map[string]*loggers.CaptureLogger, error) {
	if sinkConfig == nil {
		return kitlog.NewNopLogger(), captures, nil
	}
	numSinks := len(sinkConfig.Sinks)
	outputLoggers := make([]kitlog.Logger, numSinks, numSinks+1)
	for i, sc := range sinkConfig.Sinks {
		l, captures, err := BuildLoggerFromSinkConfig(sc, captures)
		if err != nil {
			return nil, captures, err
		}
		outputLoggers[i] = l
	}

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
	//case Syslog:
	case Stdout:
		return loggers.NewStreamLogger(os.Stdout), nil
	case Stderr:
		return loggers.NewStreamLogger(os.Stderr), nil
	case File:
		return loggers.NewFileLogger(outputConfig.FileConfig.Path)
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
			return kitlog.NewContext(outputLogger).WithPrefix(keyvals...), captures, nil
		} else {
			return kitlog.NewContext(outputLogger).With(keyvals...), captures, nil
		}
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
