package loggers

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/hyperledger/burrow/logging/structure"
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

	err = structure.Sync(fileLogger)
	require.NoError(t, err)

	bs, err := ioutil.ReadFile(logPath)

	require.NoError(t, err)
	assert.Equal(t, "{\"foo\":\"bar\"}\n", string(bs))
}

func TestFileTemplateParams(t *testing.T) {
	ftp := FileTemplateParams{
		Date: time.Now(),
	}
	fmt.Println(ftp.Timestamp())
}
