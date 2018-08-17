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
	"time"

	"sync"

	"github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/tmthrgd/go-hex"
)

// Logger that implements some formatting conventions for burrow and burrow-client
// This is intended for applying consistent value formatting before the final 'output' logger;
// we should avoid prematurely formatting values here if it is useful to let the output logger
// decide how it wants to display values. Ideal candidates for 'early' formatting here are types that
// we control and generic output loggers are unlikely to know about.
type burrowFormatLogger struct {
	sync.Mutex
	logger log.Logger
}

var _ log.Logger = &burrowFormatLogger{}

func (bfl *burrowFormatLogger) Log(keyvals ...interface{}) error {
	if bfl.logger == nil {
		return nil
	}
	if len(keyvals)%2 != 0 {
		return fmt.Errorf("log line contains an odd number of elements so "+
			"was dropped: %v", keyvals)
	}
	keyvals = structure.MapKeyValues(keyvals,
		func(key interface{}, value interface{}) (interface{}, interface{}) {
			switch v := value.(type) {
			case string:
			case fmt.Stringer:
				value = v.String()
			case []byte:
				value = hex.EncodeUpperToString(v)
			case time.Time:
				value = v.Format(time.RFC3339Nano)
			}
			return structure.StringifyKey(key), value
		})
	bfl.Lock()
	defer bfl.Unlock()
	return bfl.logger.Log(keyvals...)
}

func BurrowFormatLogger(logger log.Logger) *burrowFormatLogger {
	return &burrowFormatLogger{logger: logger}
}
