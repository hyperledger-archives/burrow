package config

import (
	"testing"
	"github.com/spf13/viper"
	"strings"
)

func TestUnmarshal(t *testing.T) {
	conf := viper.New()
	conf.ReadConfig(strings.NewReader(``))
}