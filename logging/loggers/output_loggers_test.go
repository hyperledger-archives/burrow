package loggers

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileLogger(t *testing.T) {
	f, err := ioutil.TempFile("", "TestNewFileLogger.log")
	require.NoError(t, err)
	logPath := f.Name()
	f.Close()
	fileLogger, err := NewFileLogger(logPath, JSONFormat)
	require.NoError(t, err)

	err = fileLogger.Log("foo", "bar")
	require.NoError(t, err)

	err = logging.Sync(fileLogger)
	require.NoError(t, err)

	bs, err := ioutil.ReadFile(logPath)

	require.NoError(t, err)
	assert.Equal(t, "{\"foo\":\"bar\"}\n", string(bs))
}

func TestNewStreamLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := NewStreamLogger(buf, LogfmtFormat)
	err := logger.Log("oh", "my")
	require.NoError(t, err)

	err = logging.Sync(logger)
	require.NoError(t, err)

	assert.Equal(t, "oh=my\n", string(buf.Bytes()))
}
