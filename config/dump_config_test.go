// +build dumpconfig

// Space above matters
package config

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This is a little convenience for getting a config file dump. Just run:
// go test -tags dumpconfig ./config
// This pseudo test won't run unless the dumpconfig tag is
func TestDumpConfig(t *testing.T) {
	bs, err := GetExampleConfigFileBytes()
	assert.NoError(t, err, "Should be able to create example config")
	ioutil.WriteFile("config_dump.toml", bs, 0644)
}
