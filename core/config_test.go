package core

import (
	"testing"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/definitions"
	lconfig "github.com/hyperledger/burrow/logging/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoadLoggingConfigFromDo(t *testing.T) {
	do := new(definitions.Do)
	do.Config = viper.New()
	lc, err := LoadLoggingConfigFromDo(do)
	assert.NoError(t, err)
	assert.Nil(t, lc, "Should get nil logging config when [logging] not set")
	cnf, err := config.ReadViperConfig(([]byte)(lconfig.DefaultNodeLoggingConfig().RootTOMLString()))
	assert.NoError(t, err)
	do.Config = cnf
	lc, err = LoadLoggingConfigFromDo(do)
	assert.NoError(t, err)
	assert.EqualValues(t, lconfig.DefaultNodeLoggingConfig(), lc)
}
