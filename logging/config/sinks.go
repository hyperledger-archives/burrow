package config

import (
	"fmt"
	"os"

	"regexp"

	"github.com/eapache/channels"
	"github.com/eris-ltd/eris-db/common/math/integral"
	"github.com/eris-ltd/eris-db/logging/config/types"
	"github.com/eris-ltd/eris-db/logging/loggers"
	kitlog "github.com/go-kit/kit/log"
)

func BuildLoggerFromRootSinkConfig(sinkConfig *types.SinkConfig) (kitlog.Logger, map[string]*loggers.ChannelLogger, error) {
	return BuildLoggerFromSinkConfig(sinkConfig, make(map[string]*loggers.ChannelLogger))
}

func BuildLoggerFromSinkConfig(sinkConfig *types.SinkConfig,
	captures map[string]*loggers.ChannelLogger) (kitlog.Logger, map[string]*loggers.ChannelLogger, error) {
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

	if sinkConfig.Output != nil && sinkConfig.Output.OutputType != types.NoOutput {
		l, err := BuildOutputLogger(sinkConfig.Output)
		if err != nil {
			return nil, captures, err
		}
		outputLoggers = append(outputLoggers, l)
	}

	outputLogger := loggers.NewMultipleOutputLogger(outputLoggers...)

	if sinkConfig.Transform != nil && sinkConfig.Transform.TransformType != types.NoTransform {
		return BuildTransformLogger(sinkConfig.Transform, captures, outputLogger)
	}
	return outputLogger, captures, nil
}

func BuildOutputLogger(outputConfig *types.OutputConfig) (kitlog.Logger, error) {
	switch outputConfig.OutputType {
	case types.NoOutput:
		return kitlog.NewNopLogger(), nil
	//case types.Graylog:
	//case types.Syslog:
	case types.Stdout:
		return loggers.NewStreamLogger(os.Stdout), nil
	case types.Stderr:
		return loggers.NewStreamLogger(os.Stderr), nil
	case types.File:
		return loggers.NewFileLogger(outputConfig.FileConfig.Path)
	default:
		return nil, fmt.Errorf("Could not build logger for output: '%s'", outputConfig.OutputType)
	}
}

func BuildTransformLogger(transformConfig *types.TransformConfig, captures map[string]*loggers.ChannelLogger,
	outputLogger kitlog.Logger) (kitlog.Logger, map[string]*loggers.ChannelLogger, error) {
	switch transformConfig.TransformType {
	case types.NoTransform:
		return outputLogger, captures, nil
	case types.Label:
		keyvals := make([]interface{}, 0, len(transformConfig.Labels)*2)
		for k, v := range transformConfig.LabelConfig.Labels {
			keyvals = append(keyvals, k, v)
		}
		if transformConfig.LabelConfig.Prefix {
			return kitlog.NewContext(outputLogger).WithPrefix(keyvals...), captures, nil
		} else {
			return kitlog.NewContext(outputLogger).With(keyvals...), captures, nil
		}
	case types.Capture:
		name := transformConfig.CaptureConfig.Name
		if _, ok := captures[name]; ok {
			return nil, captures, fmt.Errorf("Could not register new logging capture since name '%s' already "+
				"registered", name)
		}
		// Create a buffered channel logger to capture logs from upstream
		cl := loggers.NewChannelLogger(channels.BufferCap(transformConfig.CaptureConfig.BufferCap))
		captures[name] = cl
		// Return a logger that tees intput logs to this ChannelLogger and the passed in output logger
		return loggers.NewMultipleOutputLogger(cl, outputLogger), captures, nil
	case types.Filter:
		predicate, err := BuildFilterPredicate(transformConfig.FilterConfig)
		if err != nil {
			return nil, captures, fmt.Errorf("Could not build filter predicate: '%s'", err)
		}
		return loggers.NewFilterLogger(outputLogger, predicate), captures, nil
	default:
		return nil, captures, fmt.Errorf("Could not build logger for transform: '%s'", transformConfig.TransformType)
	}
}
