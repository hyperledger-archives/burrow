// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	. "github.com/hyperledger/burrow/logging/logconfig"
)

// Dump an example logging configuration
func main() {
	loggingConfig := &LoggingConfig{
		RootSink: Sink().
			AddSinks(
				// Log everything to Stderr
				Sink().SetOutput(StderrOutput()),
				Sink().SetTransform(FilterTransform(ExcludeWhenAllMatch,
					"module", "p2p",
					"captured_logging_source", "tendermint_log15")).
					AddSinks(
						Sink().SetOutput(StdoutOutput()),
					),
			),
	}
	fmt.Println(loggingConfig.RootTOMLString())
}
