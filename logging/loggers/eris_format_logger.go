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
	"fmt"

	"github.com/monax/burrow/logging/structure"

	kitlog "github.com/go-kit/kit/log"
)

// Logger that implements some formatting conventions for eris-db and eris-client
// This is intended for applying consistent value formatting before the final 'output' logger;
// we should avoid prematurely formatting values here if it is useful to let the output logger
// decide how it wants to display values. Ideal candidates for 'early' formatting here are types that
// we control and generic output loggers are unlikely to know about.
type erisFormatLogger struct {
	logger kitlog.Logger
}

var _ kitlog.Logger = &erisFormatLogger{}

func (efl *erisFormatLogger) Log(keyvals ...interface{}) error {
	if efl.logger == nil {
		return nil
	}
	if len(keyvals) % 2 != 0 {
		return fmt.Errorf("Log line contains an odd number of elements so " +
				"was dropped: %v", keyvals)
	}
	return efl.logger.Log(structure.MapKeyValues(keyvals, erisFormatKeyValueMapper)...)
}

func erisFormatKeyValueMapper(key, value interface{}) (interface{}, interface{}) {
	switch key {
	default:
		switch v := value.(type) {
		case []byte:
			return key, fmt.Sprintf("%X", v)
		}
	}
	return key, value
}

func ErisFormatLogger(logger kitlog.Logger) *erisFormatLogger {
	return &erisFormatLogger{logger: logger}
}
