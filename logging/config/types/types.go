package types

type Source string
type OutputType string
type TransformType string

const (
	NoOutput OutputType = ""
	Graylog  OutputType = "Graylog"
	Syslog   OutputType = "Syslog"
	File     OutputType = "File"
	Stdout   OutputType = "Stdout"
	Stderr   OutputType = "Stderr"

	NoTransform TransformType = ""
	// Filter log lines
	Filter TransformType = "Filter"
	// Remove key-val pairs from each log line
	Prune   TransformType = "Prune"
	Capture TransformType = "Capture"
	Label   TransformType = "Label"
)

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
		OutputType OutputType
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
		Name      string
		BufferCap int
	}

	// Generates true if KeyRegex matches a log line key and ValueRegex matches that key's value.
	// If ValueRegex is empty then returns true if any key matches
	// If KeyRegex is empty then returns true if any value matches
	KeyValuePredicateConfig struct {
		KeyRegex   string
		ValueRegex string
	}

	FilterConfig struct {
		// Only include log lines if they match ALL predicates in Include or include all log lines if empty
		Include []*KeyValuePredicateConfig
		// Of those log lines included by Include, exclude log lines matching ANY predicate in Exclude or include all log lines if empty
		Exclude []*KeyValuePredicateConfig
	}

	TransformConfig struct {
		TransformType TransformType
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
