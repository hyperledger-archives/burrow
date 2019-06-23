package loggers

import (
	"bytes"
	"testing"

	"github.com/hyperledger/burrow/logging/structure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStreamLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	logger, err := NewStreamLogger(buf, LogfmtFormat)
	require.NoError(t, err)
	err = logger.Log("oh", "my")
	require.NoError(t, err)

	err = structure.Sync(logger)
	require.NoError(t, err)

	assert.Equal(t, "oh=my\n", buf.String())
}

func TestNewTemplateLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	logger, err := NewTemplateLogger(buf, "Why Hello {{.name}}", []byte{'\n'})
	require.NoError(t, err)
	err = logger.Log("name", "Marjorie Stewart-Baxter", "fingertip_width_cm", float32(1.34))
	require.NoError(t, err)
	err = logger.Log("name", "Fred")
	require.NoError(t, err)
	assert.Equal(t, "Why Hello Marjorie Stewart-Baxter\nWhy Hello Fred\n", buf.String())
}
