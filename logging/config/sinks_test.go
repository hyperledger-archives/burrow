package config

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildLoggerFromSinkConfig(t *testing.T) {
	sinkConfig := Sink().
		AddSinks(
			Sink().
				AddSinks(
					Sink().
						AddSinks(
							Sink().
								SetTransform(CaptureTransform("cap", 100, true)).
								SetOutput(StderrOutput()).
								AddSinks(
									Sink().
										SetTransform(LabelTransform(true, "Label", "A Label!")).
										SetOutput(StdoutOutput())))))

	logger, captures, err := sinkConfig.BuildLogger()
	logger.Log("Foo", "Bar")
	assert.NoError(t, err)
	assert.Equal(t, logLines("Foo", "Bar"),
		captures["cap"].BufferLogger().FlushLogLines())
}

func TestFileLoggerSink(t *testing.T) {
	sinkConfig := Sink().
		SetOutput(FileOutput("/tmp/logmclogface")).AddSinks(
		Sink().SetOutput(FileOutput("/tmp/doubleloglog")))

	bs, err := json.Marshal(sinkConfig)
	assert.NoError(t, err)

	fmt.Println(string(bs))
	_, _, err = sinkConfig.BuildLogger()
	assert.NoError(t, err)
}

func TestFilterSinks(t *testing.T) {
	sinkConfig := Sink().
		SetOutput(StderrOutput()).
		AddSinks(
			Sink().
				SetTransform(FilterTransform(IncludeWhenAnyMatches,
					"Foo", "Bar",
					"Rough", "Trade",
				)).
				AddSinks(
					Sink().
						SetTransform(CaptureTransform("Included", 100, true)).
						AddSinks(
							Sink().
								SetTransform(FilterTransform(ExcludeWhenAllMatch,
									"Foo", "Baz",
									"Index", "00$")).
								AddSinks(
									Sink().
										SetTransform(CaptureTransform("Excluded", 100, false)),
								),
						),
				),
		)

	logger, captures, err := sinkConfig.BuildLogger()
	assert.NoError(t, err, "Should be able to build filter logger")
	included := captures["Included"]
	excluded := captures["Excluded"]

	// Included by both filters
	ll := logLines("Foo", "Bar")
	logger.Log(ll[0]...)
	assert.Equal(t, logLines("Foo", "Bar"),
		included.BufferLogger().FlushLogLines())
	assert.Equal(t, logLines("Foo", "Bar"),
		excluded.BufferLogger().FlushLogLines())

	// Included by first filter and excluded by second
	ll = logLines("Foo", "Bar", "Foo", "Baz", "Index", "1000")
	logger.Log(ll[0]...)
	assert.Equal(t, ll, included.BufferLogger().FlushLogLines())
	assert.Equal(t, logLines(), excluded.BufferLogger().FlushLogLines())

	// Included by first filter and not excluded by second despite matching one
	// predicate
	ll = logLines("Rough", "Trade", "Index", "1000")
	logger.Log(ll[0]...)
	assert.Equal(t, ll, included.BufferLogger().FlushLogLines())
	assert.Equal(t, ll, excluded.BufferLogger().FlushLogLines())
}

func TestPruneTransform(t *testing.T) {
	sinkConfig := Sink().
		SetTransform(PruneTransform("Trace")).
		AddSinks(Sink().
			SetTransform(CaptureTransform("cap", 100, false)))

	logger, captures, err := sinkConfig.BuildLogger()
	assert.NoError(t, err)
	logger.Log("msg", "Hello with a trace",
		"Trace", []string{"logger:32, state:23"})
	logger.Log("msg", "Goodbye with a trace",
		"Trace", []string{"logger:32, state:14"})
	assert.Equal(t, logLines("msg", "Hello with a trace", "",
		"msg", "Goodbye with a trace"),
		captures["cap"].FlushLogLines())
}

// Takes a variadic argument of log lines as a list of key value pairs delimited
// by the empty string
func logLines(keyvals ...string) [][]interface{} {
	llines := make([][]interface{}, 0)
	line := make([]interface{}, 0)
	for _, kv := range keyvals {
		if kv == "" {
			llines = append(llines, line)
			line = make([]interface{}, 0)
		} else {
			line = append(line, kv)
		}
	}
	if len(line) > 0 {
		llines = append(llines, line)
	}
	return llines
}
