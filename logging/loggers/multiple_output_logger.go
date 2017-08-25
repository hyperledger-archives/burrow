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

package loggers

import (
	"strings"

	kitlog "github.com/go-kit/kit/log"
)

// This represents an 'AND' type logger. When logged to it will log to each of
// the loggers in the slice.
type MultipleOutputLogger []kitlog.Logger

var _ kitlog.Logger = MultipleOutputLogger(nil)

func (mol MultipleOutputLogger) Log(keyvals ...interface{}) error {
	var errs []error
	for _, logger := range mol {
		err := logger.Log(keyvals...)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return combineErrors(errs)
}

// Creates a logger that forks log messages to each of its outputLoggers
func NewMultipleOutputLogger(outputLoggers ...kitlog.Logger) kitlog.Logger {
	moLogger := make(MultipleOutputLogger, 0, len(outputLoggers))
	// Flatten any MultipleOutputLoggers
	for _, ol := range outputLoggers {
		if ls, ok := ol.(MultipleOutputLogger); ok {
			moLogger = append(moLogger, ls...)
		} else {
			moLogger = append(moLogger, ol)
		}
	}
	return moLogger
}

type multipleErrors []error

func combineErrors(errs []error) error {
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return multipleErrors(errs)
	}
}

func (errs multipleErrors) Error() string {
	var errStrings []string
	for _, err := range errs {
		errStrings = append(errStrings, err.Error())
	}
	return strings.Join(errStrings, ";")
}
