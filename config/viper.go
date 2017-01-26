package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Safely get the subtree from a viper config, returning an error if it could not
// be obtained for any reason.
func ViperSubConfig(conf *viper.Viper, configSubtreePath string) (subConfig *viper.Viper, err error) {
	// Viper internally panics if `moduleName` contains an unallowed
	// character (eg, a dash).
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Viper panicked trying to read config subtree: %s",
				configSubtreePath)
		}
	}()
	if !conf.IsSet(configSubtreePath) {
		return nil, fmt.Errorf("Failed to read config subtree: %s",
			configSubtreePath)
	}
	subConfig = conf.Sub(configSubtreePath)
	if subConfig == nil {
		return nil, fmt.Errorf("Failed to read config subtree: %s",
			configSubtreePath)
	}
	return subConfig, err
}
