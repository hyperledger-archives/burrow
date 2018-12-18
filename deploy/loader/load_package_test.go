package loader

import (
	"bytes"
	"testing"

	"github.com/hyperledger/burrow/deploy/def"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestUnmarshal(t *testing.T) {
	testUnmarshal(t, `jobs:

- name: AddValidators
  update-account:
    source: foo
    target: bar
    permissions: [foo, bar]
    roles: []

- name: nameRegTest1
  register:
    name: $val1
    data: $val2
    amount: $to_save
    fee: $MinersFee
`)
	testUnmarshal(t, `jobs:

  update-account:
    source: foo
    target: bar
    permissions: [foo, bar]
    roles: []
`)
}

func testUnmarshal(t *testing.T, testPackageYAML string) {
	pkgs := viper.New()
	pkgs.SetConfigType("yaml")
	err := pkgs.ReadConfig(bytes.NewBuffer([]byte(testPackageYAML)))
	require.NoError(t, err)
	do := new(def.Package)

	err = pkgs.UnmarshalExact(do)
	require.NoError(t, err)
	yamlOut, err := yaml.Marshal(do)
	require.NoError(t, err)
	assert.True(t, len(yamlOut) > 100, "should marshal some yaml")

	doOut := new(def.Package)
	err = yaml.Unmarshal(yamlOut, doOut)
	require.NoError(t, err)
	assert.Equal(t, do, doOut)
}
