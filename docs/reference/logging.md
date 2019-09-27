# Logging

Logging is highly configurable through the `burrow.toml` `[logging]` section. Each log line is a list of key-value pairs that flows from the root sink through possible child sinks. 
Each sink can have an output, a transform, and sinks that it outputs to. Below is a more involved example than the one appearing in the default generated config of what you can configure:

```toml
# This is a top level config section within the main Burrow config
[logging]
  # All log lines are sent to the root sink from all sources
  [logging.root_sink]
    # We define two child sinks that each receive all log lines
    [[logging.root_sink.sinks]]
      # We send all output to stderr
      [logging.root_sink.sinks.output]
        output_type = "stderr"

    [[logging.root_sink.sinks]]
      # But for the second sink we define a transform that filters log lines from Tendermint's p2p module
      [logging.root_sink.sinks.transform]
        transform_type = "filter"
        filter_mode = "exclude_when_all_match"

        [[logging.root_sink.sinks.transform.predicates]]
          key_regex = "module"
          value_regex = "p2p"

        [[logging.root_sink.sinks.transform.predicates]]
          key_regex = "captured_logging_source"
          value_regex = "tendermint_log15"

      # The child sinks of this filter transform sink are syslog and file and will omit log lines originating from p2p
      [[logging.root_sink.sinks.sinks]]
        [logging.root_sink.sinks.sinks.output]
          output_type = "syslog"
          url = ""
          tag = "Burrow-network"

      [[logging.root_sink.sinks.sinks]]
        [logging.root_sink.sinks.sinks.output]
          output_type = "file"
          path = "/var/log/burrow-network.log"
```