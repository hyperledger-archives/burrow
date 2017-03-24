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
	"runtime"
	"testing"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/monax/eris-db/logging/structure"
	"github.com/stretchr/testify/assert"
)

func TestMultipleChannelLogger(t *testing.T) {
	boringLogger, interestingLogger := newTestLogger(), newTestLogger()
	mcl := kitlog.NewContext(MultipleChannelLogger(map[string]kitlog.Logger{
		"Boring":      boringLogger,
		"Interesting": interestingLogger,
	}))
	err := mcl.With("time", kitlog.Valuer(func() interface{} { return "aa" })).
		Log(structure.ChannelKey, "Boring", "foo", "bar")
	assert.NoError(t, err, "Should log without an error")
	// Wait for channel to drain
	time.Sleep(time.Second)
	runtime.Gosched()
	assert.Equal(t, []interface{}{"time", "aa", structure.ChannelKey, "Boring",
		"foo", "bar"},
		boringLogger.logLines[0])
}
