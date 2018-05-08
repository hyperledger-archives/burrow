package source

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironment(t *testing.T) {
	envVar := "FISH_SPOON_ALPHA"
	jsonString := JSONString(newTestConfig())
	os.Setenv(envVar, jsonString)
	conf := new(animalConfig)
	err := Environment(envVar).Apply(conf)
	assert.NoError(t, err)
	assert.Equal(t, jsonString, JSONString(conf))
}

func TestDeepCopy(t *testing.T) {
	conf := newTestConfig()
	confCopy, err := DeepCopy(conf)
	require.NoError(t, err)
	assert.Equal(t, conf, confCopy)
}

func TestFile(t *testing.T) {
	tomlString := TOMLString(newTestConfig())
	file := writeConfigFile(t, newTestConfig())
	defer os.Remove(file)
	conf := new(animalConfig)
	err := File(file, false).Apply(conf)
	assert.NoError(t, err)
	assert.Equal(t, tomlString, TOMLString(conf))
}

func TestCascade(t *testing.T) {
	envVar := "FISH_SPOON_ALPHA"
	// Both fall through so baseConfig returned
	conf := newTestConfig()
	err := Cascade(os.Stderr, true,
		Environment(envVar),
		File("", false)).Apply(conf)
	assert.NoError(t, err)
	assert.Equal(t, newTestConfig(), conf)

	// Env not set so falls through to file
	fileConfig := newTestConfig()
	file := writeConfigFile(t, fileConfig)
	defer os.Remove(file)
	conf = new(animalConfig)
	err = Cascade(os.Stderr, true,
		Environment(envVar),
		File(file, false)).Apply(conf)
	assert.NoError(t, err)
	assert.Equal(t, TOMLString(fileConfig), TOMLString(conf))

	// Env set so caught by environment source
	envConfig := animalConfig{
		Name:    "Slug",
		NumLegs: 0,
	}
	os.Setenv(envVar, JSONString(envConfig))
	conf = newTestConfig()
	err = Cascade(os.Stderr, true,
		Environment(envVar),
		File(file, false)).Apply(conf)
	assert.NoError(t, err)
	assert.Equal(t, TOMLString(envConfig), TOMLString(conf))
}

func TestDetectFormat(t *testing.T) {
	assert.Equal(t, TOML, DetectFormat(""))
	assert.Equal(t, JSON, DetectFormat("{"))
	assert.Equal(t, JSON, DetectFormat("\n\n\t    \n\n      {"))
	assert.Equal(t, TOML, DetectFormat("[Tendermint]\n  Seeds =\"foobar@val0\"}"))
}

func writeConfigFile(t *testing.T, conf interface{}) string {
	tomlString := TOMLString(conf)
	f, err := ioutil.TempFile("", "source-test.toml")
	assert.NoError(t, err)
	f.Write(([]byte)(tomlString))
	f.Close()
	return f.Name()
}

// Test types

type legConfig struct {
	Leg    int
	Colour byte
}

type animalConfig struct {
	Name    string
	NumLegs int
	Legs    []legConfig
}

func newTestConfig() *animalConfig {
	return &animalConfig{
		Name:    "Froggy!",
		NumLegs: 2,
		Legs: []legConfig{
			{
				Leg:    1,
				Colour: 034,
			},
			{
				Leg:    2,
				Colour: 034,
			},
		},
	}
}
