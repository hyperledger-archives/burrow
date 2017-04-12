package config

import (
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	. "github.com/tendermint/go-common"
)

type Config interface {
	Get(key string) interface{}
	GetBool(key string) bool
	GetFloat64(key string) float64
	GetInt(key string) int
	GetString(key string) string
	GetStringSlice(key string) []string
	GetTime(key string) time.Time
	GetMap(key string) map[string]interface{}
	GetMapString(key string) map[string]string
	GetConfig(key string) Config
	IsSet(key string) bool
	Set(key string, value interface{})
	SetDefault(key string, value interface{})
}

type MapConfig struct {
	mtx      sync.Mutex
	required map[string]struct{} // blows up if trying to use before setting.
	data     map[string]interface{}
}

func ReadMapConfigFromFile(filePath string) (*MapConfig, error) {
	var configData = make(map[string]interface{})
	fileBytes := MustReadFile(filePath)
	err := toml.Unmarshal(fileBytes, &configData)
	if err != nil {
		return nil, err
	}
	return NewMapConfig(configData), nil
}

func NewMapConfig(data map[string]interface{}) *MapConfig {
	if data == nil {
		data = make(map[string]interface{})
	}
	return &MapConfig{
		required: make(map[string]struct{}),
		data:     data,
	}
}

func (cfg *MapConfig) Get(key string) interface{} {
	cfg.mtx.Lock()
	defer cfg.mtx.Unlock()
	if _, ok := cfg.required[key]; ok {
		PanicSanity(Fmt("config key %v is required but was not set.", key))
	}
	spl := strings.Split(key, ".")
	l := len(spl)
	if l > 1 {
		first, keyPath, keyBase := spl[0], spl[1:l-1], spl[l-1]

		f := cfg.data[first].(map[string]interface{})
		for _, k := range keyPath {
			f = f[k].(map[string]interface{})
		}
		return f[keyBase]
	}
	return cfg.data[key]
}
func (cfg *MapConfig) GetBool(key string) bool       { return cfg.Get(key).(bool) }
func (cfg *MapConfig) GetFloat64(key string) float64 { return cfg.Get(key).(float64) }
func (cfg *MapConfig) GetInt(key string) int {
	switch v := cfg.Get(key).(type) {
	case int:
		return v
	case int64:
		// when loaded from toml file, ints come as int64
		return int(v)
	}
	return cfg.Get(key).(int) // panic
}
func (cfg *MapConfig) GetString(key string) string { return cfg.Get(key).(string) }
func (cfg *MapConfig) GetMap(key string) map[string]interface{} {
	return cfg.Get(key).(map[string]interface{})
}
func (cfg *MapConfig) GetMapString(key string) map[string]string {
	return cfg.Get(key).(map[string]string)
}
func (cfg *MapConfig) GetConfig(key string) Config {
	v := cfg.Get(key)
	if v == nil {
		return NewMapConfig(nil)
	}
	return NewMapConfig(v.(map[string]interface{}))
}
func (cfg *MapConfig) GetStringSlice(key string) []string { return cfg.Get(key).([]string) }
func (cfg *MapConfig) GetTime(key string) time.Time       { return cfg.Get(key).(time.Time) }
func (cfg *MapConfig) IsSet(key string) bool {
	cfg.mtx.Lock()
	defer cfg.mtx.Unlock()
	_, ok := cfg.data[key]
	return ok
}
func (cfg *MapConfig) Set(key string, value interface{}) {
	cfg.mtx.Lock()
	defer cfg.mtx.Unlock()
	delete(cfg.required, key)
	cfg.set(key, value)
}
func (cfg *MapConfig) set(key string, value interface{}) {
	spl := strings.Split(key, ".")
	l := len(spl)
	if l > 1 {
		first, keyPath, keyBase := spl[0], spl[1:l-1], spl[l-1]

		f := assertOrNewMap(cfg.data, first)
		for _, k := range keyPath {
			f = assertOrNewMap(f, k)
		}
		f[keyBase] = value
		return
	}
	cfg.data[key] = value
}

func (cfg *MapConfig) SetDefault(key string, value interface{}) {
	cfg.mtx.Lock()
	delete(cfg.required, key)
	cfg.mtx.Unlock()

	if cfg.IsSet(key) {
		return
	}

	cfg.mtx.Lock()
	cfg.set(key, value)
	cfg.mtx.Unlock()
}

func (cfg *MapConfig) SetRequired(key string) {
	if cfg.IsSet(key) {
		return
	}
	cfg.mtx.Lock()
	cfg.required[key] = struct{}{}
	cfg.mtx.Unlock()
}

func assertOrNewMap(dataMap map[string]interface{}, key string) map[string]interface{} {
	m, ok := dataMap[key].(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		dataMap[key] = m
	}
	return m
}
