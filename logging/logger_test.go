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

package logging

import (
	"bytes"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	bufLogger := log.NewLogfmtLogger(&buf)
	logger := NewLogger(bufLogger)
	logger = logger.With("foo", "bar")
	logger.Trace.Log("hello", "barry")
	require.Equal(t, "log_channel=Trace foo=bar hello=barry\n", buf.String())
}

func TestNewNoopInfoTraceLogger(t *testing.T) {
	logger := NewNoopLogger()
	logger.Trace.Log("goodbye", "trevor")
}
