// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stdlib

import (
	"io"
	"log"

	kitlog "github.com/go-kit/kit/log"
	"github.com/monax/eris-db/logging/loggers"
)

func Capture(stdLibLogger log.Logger,
	logger loggers.InfoTraceLogger) io.Writer {
	adapter := newAdapter(logger)
	stdLibLogger.SetOutput(adapter)
	return adapter
}

func CaptureRootLogger(logger loggers.InfoTraceLogger) io.Writer {
	adapter := newAdapter(logger)
	log.SetOutput(adapter)
	return adapter
}

func newAdapter(logger loggers.InfoTraceLogger) io.Writer {
	return kitlog.NewStdlibAdapter(logger)
}
