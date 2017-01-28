package logging

type (
	SinkConfig struct {
		Channels []string
	}

	LoggingConfig struct {
		Sinks []SinkConfig
	}
)
