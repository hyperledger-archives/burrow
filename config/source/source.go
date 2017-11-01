package source

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/cep21/xdgbasedir"
	"github.com/imdario/mergo"
)

// If passed this identifier try to read config from STDIN
const STDINFileIdentifier = "-"

type ConfigProvider interface {
	// Description of where this provider sources its config from
	From() string
	// Get the config values to the passed in baseConfig
	Apply(baseConfig interface{}) error
	// Return a copy of the provider that does nothing if skip is true
	SetSkip(skip bool) ConfigProvider
	// Whether to skip this provider
	Skip() bool
}

var _ ConfigProvider = &configSource{}

type configSource struct {
	from  string
	skip  bool
	apply func(baseConfig interface{}) error
}

func NewConfigProvider(from string, skip bool, apply func(baseConfig interface{}) error) *configSource {
	return &configSource{
		from:  from,
		skip:  skip,
		apply: apply,
	}
}

func (cs *configSource) From() string {
	return cs.from
}

func (cs *configSource) Apply(baseConfig interface{}) error {
	return cs.apply(baseConfig)
}

func (cs *configSource) Skip() bool {
	return cs.skip
}

// Returns a copy of the configSource with skip set as passed in
func (cs *configSource) SetSkip(skip bool) ConfigProvider {
	return &configSource{
		skip:  skip,
		from:  cs.from,
		apply: cs.apply,
	}
}

// Builds a ConfigProvider by iterating over a cascade of ConfigProvider sources. Can be used
// in two distinct modes: with shortCircuit true the first successful ConfigProvider source
// is returned. With shortCircuit false sources appearing later are used to possibly override
// those appearing earlier
func Cascade(logWriter io.Writer, shortCircuit bool, providers ...ConfigProvider) *configSource {
	var fromStrings []string
	skip := true
	for _, provider := range providers {
		if !provider.Skip() {
			skip = false
			fromStrings = append(fromStrings, provider.From())
		}
	}
	fromPrefix := "each of"
	if shortCircuit {
		fromPrefix = "first of"

	}
	return &configSource{
		skip: skip,
		from: fmt.Sprintf("%s: %s", fromPrefix, strings.Join(fromStrings, " then ")),
		apply: func(baseConfig interface{}) error {
			if baseConfig == nil {
				return fmt.Errorf("baseConfig passed to Cascade(...).Get() must not be nil")
			}
			for _, provider := range providers {
				if !provider.Skip() {
					writeLog(logWriter, fmt.Sprintf("Sourcing config from %s", provider.From()))
					err := provider.Apply(baseConfig)
					if err != nil {
						return err
					}
					if shortCircuit {
						return nil
					}
				}
			}
			return nil
		},
	}
}

func FirstOf(providers ...ConfigProvider) *configSource {
	return Cascade(os.Stderr, true, providers...)
}

func EachOf(providers ...ConfigProvider) *configSource {
	return Cascade(os.Stderr, false, providers...)
}

// Try to source config from provided JSON file, is skipNonExistent is true then the provider will fall-through (skip)
// when the file doesn't exist, rather than returning an error
func JSONFile(configFile string, skipNonExistent bool) *configSource {
	return &configSource{
		skip: ShouldSkipFile(configFile, skipNonExistent),
		from: fmt.Sprintf("JSON config file at '%s'", configFile),
		apply: func(baseConfig interface{}) error {
			return FromJSONFile(configFile, baseConfig)
		},
	}
}

// Try to source config from provided TOML file, is skipNonExistent is true then the provider will fall-through (skip)
// when the file doesn't exist, rather than returning an error
func TOMLFile(configFile string, skipNonExistent bool) *configSource {
	return &configSource{
		skip: ShouldSkipFile(configFile, skipNonExistent),
		from: fmt.Sprintf("TOML config file at '%s'", configFile),
		apply: func(baseConfig interface{}) error {
			return FromTOMLFile(configFile, baseConfig)
		},
	}
}

// Try to find config by using XDG base dir spec
func XDGBaseDir(configFileName string) *configSource {
	skip := false
	// Look for config in standard XDG specified locations
	configFile, err := xdgbasedir.GetConfigFileLocation(configFileName)
	if err == nil {
		_, err := os.Stat(configFile)
		// Skip if config  file does not exist at default location
		skip = os.IsNotExist(err)
	}
	return &configSource{
		skip: skip,
		from: fmt.Sprintf("XDG base dir"),
		apply: func(baseConfig interface{}) error {
			if err != nil {
				return err
			}
			return FromTOMLFile(configFile, baseConfig)
		},
	}
}

// Source from a single environment variable with config embedded in JSON
func Environment(key string) *configSource {
	jsonString := os.Getenv(key)
	return &configSource{
		skip: jsonString == "",
		from: fmt.Sprintf("'%s' environment variable (as JSON)", key),
		apply: func(baseConfig interface{}) error {
			return FromJSONString(jsonString, baseConfig)
		},
	}
}

func Default(defaultConfig interface{}) *configSource {
	return &configSource{
		from: "defaults",
		apply: func(baseConfig interface{}) error {
			return mergo.MergeWithOverwrite(baseConfig, defaultConfig)
		},
	}
}

func FromJSONFile(configFile string, conf interface{}) error {
	bs, err := ReadFile(configFile)
	if err != nil {
		return err
	}

	return FromJSONString(string(bs), conf)
}

func FromTOMLFile(configFile string, conf interface{}) error {
	bs, err := ReadFile(configFile)
	if err != nil {
		return err
	}

	return FromTOMLString(string(bs), conf)
}

func FromTOMLString(tomlString string, conf interface{}) error {
	_, err := toml.Decode(tomlString, conf)
	if err != nil {
		return err
	}
	return nil
}

func FromJSONString(jsonString string, conf interface{}) error {
	err := json.Unmarshal(([]byte)(jsonString), conf)
	if err != nil {
		return err
	}
	return nil
}

func TOMLString(conf interface{}) string {
	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(buf)
	err := encoder.Encode(conf)
	if err != nil {
		return fmt.Sprintf("<Could not serialise config>")
	}
	return buf.String()
}

func JSONString(conf interface{}) string {
	bs, err := json.MarshalIndent(conf, "", "\t")
	if err != nil {
		return fmt.Sprintf("<Could not serialise config>")
	}
	return string(bs)
}

func Merge(base, override interface{}) (interface{}, error) {
	merged, err := DeepCopy(base)
	if err != nil {
		return nil, err
	}
	err = mergo.MergeWithOverwrite(merged, override)
	if err != nil {
		return nil, err
	}
	return merged, nil
}

// Passed a pointer to struct creates a deep copy of the struct
func DeepCopy(conf interface{}) (interface{}, error) {
	// Create a zero value
	confCopy := reflect.New(reflect.TypeOf(conf).Elem()).Interface()
	// Perform a merge into that value to effect the copy
	err := mergo.Merge(confCopy, conf)
	if err != nil {
		return nil, err
	}
	return confCopy, nil
}

func writeLog(writer io.Writer, msg string) {
	if writer != nil {
		writer.Write(([]byte)(msg))
		writer.Write(([]byte)("\n"))
	}
}

func ReadFile(file string) ([]byte, error) {
	if file == STDINFileIdentifier {
		return ioutil.ReadAll(os.Stdin)
	}
	return ioutil.ReadFile(file)
}

func ShouldSkipFile(file string, skipNonExistent bool) bool {
	skip := file == ""
	if !skip && skipNonExistent {
		_, err := os.Stat(file)
		skip = os.IsNotExist(err)
	}
	return skip
}
