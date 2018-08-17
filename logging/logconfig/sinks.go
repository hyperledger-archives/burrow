package logconfig

import (
	"fmt"
	"os"

	"github.com/eapache/channels"
	"github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/hyperledger/burrow/logging/structure"
)

// This file contains definitions for a configurable output graph for the
// logging system.

type outputType string
type transformType string
type filterMode string

const (
	// OutputType
	NoOutput outputType = ""
	Stdout   outputType = "stdout"
	Stderr   outputType = "stderr"
	File     outputType = "file"

	// TransformType
	NoTransform transformType = ""
	// Filter log lines
	Filter transformType = "filter"
	// Remove key-val pairs from each log line
	Prune transformType = "prune"
	// Add key value pairs to each log line
	Label     transformType = "label"
	Capture   transformType = "capture"
	Sort      transformType = "sort"
	Vectorise transformType = "vectorise"

	// TODO [Silas]: add 'flush on exit' transform which flushes the buffer of
	// CaptureLogger to its OutputLogger a non-passthrough capture when an exit
	// signal is detected or some other exceptional thing happens

	NoFilterMode          filterMode = ""
	IncludeWhenAllMatch   filterMode = "include_when_all_match"
	IncludeWhenAnyMatches filterMode = "include_when_any_match"
	ExcludeWhenAllMatch   filterMode = "exclude_when_all_match"
	ExcludeWhenAnyMatches filterMode = "exclude_when_any_match"
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
	// TODO: reintroduce syslog removed when we dropped log15 dependency
	SyslogConfig struct {
		Url string
		Tag string
	}

	FileConfig struct {
		Path string
	}

	OutputConfig struct {
		OutputType    outputType
		Format        string
		*FileConfig   `json:",omitempty" toml:",omitempty"`
		*SyslogConfig `json:",omitempty" toml:",omitempty"`
	}

	// Transforms
	LabelConfig struct {
		Labels map[string]string
		Prefix bool
	}

	PruneConfig struct {
		Keys        []string
		IncludeKeys bool
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

	SortConfig struct {
		// Sort keys-values with keys in this list first
		Keys []string
	}

	TransformConfig struct {
		TransformType transformType
		LabelConfig   *LabelConfig   `json:",omitempty" toml:",omitempty"`
		PruneConfig   *PruneConfig   `json:",omitempty" toml:",omitempty"`
		CaptureConfig *CaptureConfig `json:",omitempty" toml:",omitempty"`
		FilterConfig  *FilterConfig  `json:",omitempty" toml:",omitempty"`
		SortConfig    *SortConfig    `json:",omitempty" toml:",omitempty"`
	}

	// Sink
	// A Sink describes a logger that logs to zero or one output and logs to zero or more child sinks.
	// before transmitting its log it applies zero or one transforms to the stream of log lines.
	// by chaining together many Sinks arbitrary transforms to and multi
	SinkConfig struct {
		Transform *TransformConfig `json:",omitempty" toml:",omitempty"`
		Sinks     []*SinkConfig    `json:",omitempty" toml:",omitempty"`
		Output    *OutputConfig    `json:",omitempty" toml:",omitempty"`
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

func (outputConfig *OutputConfig) SetFormat(format string) *OutputConfig {
	outputConfig.Format = format
	return outputConfig
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

func FileOutput(path string) *OutputConfig {
	return &OutputConfig{
		OutputType: File,
		FileConfig: &FileConfig{
			Path: path,
		},
	}
}

// Transforms

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

func OnlyTransform(keys ...string) *TransformConfig {
	return &TransformConfig{
		TransformType: Prune,
		PruneConfig: &PruneConfig{
			Keys:        keys,
			IncludeKeys: true,
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

func FilterTransform(fmode filterMode, keyValueRegexes ...string) *TransformConfig {
	if len(keyValueRegexes)%2 == 1 {
		keyValueRegexes = append(keyValueRegexes, "")
	}
	length := len(keyValueRegexes) / 2
	predicates := make([]*KeyValuePredicateConfig, length)
	for i := 0; i < length; i++ {
		kv := i * 2
		predicates[i] = &KeyValuePredicateConfig{
			KeyRegex:   keyValueRegexes[kv],
			ValueRegex: keyValueRegexes[kv+1],
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

func (filterConfig *FilterConfig) SetFilterMode(filterMode filterMode) *FilterConfig {
	filterConfig.FilterMode = filterMode
	return filterConfig
}

func (filterConfig *FilterConfig) AddPredicate(keyRegex, valueRegex string) *FilterConfig {
	filterConfig.Predicates = append(filterConfig.Predicates, &KeyValuePredicateConfig{
		KeyRegex:   keyRegex,
		ValueRegex: valueRegex,
	})
	return filterConfig
}

func SortTransform(keys ...string) *TransformConfig {
	return &TransformConfig{
		TransformType: Sort,
		SortConfig: &SortConfig{
			Keys: keys,
		},
	}
}

func VectoriseTransform() *TransformConfig {
	return &TransformConfig{
		TransformType: Vectorise,
	}
}

// Logger formation
func (sinkConfig *SinkConfig) BuildLogger() (log.Logger, map[string]*loggers.CaptureLogger, error) {
	return BuildLoggerFromSinkConfig(sinkConfig, make(map[string]*loggers.CaptureLogger))
}

func BuildLoggerFromSinkConfig(sinkConfig *SinkConfig, captures map[string]*loggers.CaptureLogger) (log.Logger,
	map[string]*loggers.CaptureLogger, error) {

	if sinkConfig == nil {
		return log.NewNopLogger(), captures, nil
	}
	numSinks := len(sinkConfig.Sinks)
	outputLoggers := make([]log.Logger, numSinks, numSinks+1)
	// We need a depth-first post-order over the output loggers so we'll keep
	// recurring into children sinks we reach a terminal sink (with no children)
	for i, sc := range sinkConfig.Sinks {
		var err error
		outputLoggers[i], captures, err = BuildLoggerFromSinkConfig(sc, captures)
		if err != nil {
			return nil, nil, err
		}
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
		return BuildTransformLoggerPassthrough(sinkConfig.Transform, captures, outputLogger)
	}
	return outputLogger, captures, nil
}

func BuildOutputLogger(outputConfig *OutputConfig) (log.Logger, error) {
	switch outputConfig.OutputType {
	case NoOutput:
		return log.NewNopLogger(), nil
	case File:
		return loggers.NewFileLogger(outputConfig.FileConfig.Path, outputConfig.Format)
	case Stdout:
		return loggers.NewStreamLogger(os.Stdout, outputConfig.Format)
	case Stderr:
		return loggers.NewStreamLogger(os.Stderr, outputConfig.Format)
	default:
		return nil, fmt.Errorf("could not build logger for output: '%s'",
			outputConfig.OutputType)
	}
}

func BuildTransformLoggerPassthrough(transformConfig *TransformConfig, captures map[string]*loggers.CaptureLogger,
	outputLogger log.Logger) (log.Logger, map[string]*loggers.CaptureLogger, error) {

	transformThenOutputLogger, captures, err := BuildTransformLogger(transformConfig, captures, outputLogger)
	if err != nil {
		return nil, nil, err
	}
	// send signals through captures so they can be flushed
	if transformConfig.TransformType == Capture {
		return transformThenOutputLogger, captures, nil
	}
	return signalPassthroughLogger(outputLogger, transformThenOutputLogger), captures, nil
}

func BuildTransformLogger(transformConfig *TransformConfig, captures map[string]*loggers.CaptureLogger,
	outputLogger log.Logger) (log.Logger, map[string]*loggers.CaptureLogger, error) {

	switch transformConfig.TransformType {
	case NoTransform:
		return outputLogger, captures, nil
	case Label:
		if transformConfig.LabelConfig == nil {
			return nil, nil, fmt.Errorf("label transform specified but no LabelConfig provided")
		}
		keyvals := make([]interface{}, 0, len(transformConfig.LabelConfig.Labels)*2)
		for k, v := range transformConfig.LabelConfig.Labels {
			keyvals = append(keyvals, k, v)
		}
		if transformConfig.LabelConfig.Prefix {
			return log.WithPrefix(outputLogger, keyvals...), captures, nil
		} else {
			return log.With(outputLogger, keyvals...), captures, nil
		}
	case Prune:
		if transformConfig.PruneConfig == nil {
			return nil, nil, fmt.Errorf("prune transform specified but no PruneConfig provided")
		}
		keys := make([]interface{}, len(transformConfig.PruneConfig.Keys))
		for i, k := range transformConfig.PruneConfig.Keys {
			keys[i] = k
		}
		return log.LoggerFunc(func(keyvals ...interface{}) error {
			if transformConfig.PruneConfig.IncludeKeys {
				return outputLogger.Log(structure.OnlyKeys(keyvals, keys...)...)
			} else {
				return outputLogger.Log(structure.RemoveKeys(keyvals, keys...)...)
			}
		}), captures, nil

	case Capture:
		if transformConfig.CaptureConfig == nil {
			return nil, nil, fmt.Errorf("capture transform specified but no CaptureConfig provided")
		}
		name := transformConfig.CaptureConfig.Name
		if _, ok := captures[name]; ok {
			return nil, captures, fmt.Errorf("could not register new logging capture since name '%s' already "+
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
		if transformConfig.FilterConfig == nil {
			return nil, nil, fmt.Errorf("filter transform specified but no FilterConfig provided")
		}
		predicate, err := BuildFilterPredicate(transformConfig.FilterConfig)
		if err != nil {
			return nil, captures, fmt.Errorf("could not build filter predicate: '%s'", err)
		}
		return loggers.FilterLogger(outputLogger, predicate), captures, nil
	case Sort:
		if transformConfig.SortConfig == nil {
			return nil, nil, fmt.Errorf("sort transform specified but no SortConfig provided")
		}
		return loggers.SortLogger(outputLogger, transformConfig.SortConfig.Keys...), captures, nil
	case Vectorise:
		return loggers.VectorValuedLogger(outputLogger), captures, nil
	default:
		return nil, captures, fmt.Errorf("could not build logger for transform: '%s'", transformConfig.TransformType)
	}
}

func signalPassthroughLogger(ifSignalLogger log.Logger, otherwiseLogger log.Logger) log.Logger {
	return log.LoggerFunc(func(keyvals ...interface{}) error {
		if structure.Signal(keyvals) != "" {
			return ifSignalLogger.Log(keyvals...)
		}
		return otherwiseLogger.Log(keyvals...)
	})
}
