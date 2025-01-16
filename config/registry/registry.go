package registry

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

var (
	mu      sync.RWMutex
	configs = make(map[string]Config)
)

// IteratorFunc is a function that is used to iterate through all registered configs.
//
// The iterator should return `true` if it wants to move to the next config, and `false` otherwise.
type IteratorFunc func(name string, config Config) bool

// RegisterConfig registers a Config with given name.
func RegisterConfig(name string, config Config) {
	mu.Lock()
	defer mu.Unlock()

	_, ok := configs[name]
	if ok {
		panic(fmt.Sprintf("a config with name %s already exist", name))
	}

	configs[name] = config
}

// GetConfig returns a Config with given name.
func GetConfig(name string) Config {
	mu.RLock()
	defer mu.RUnlock()

	config, ok := configs[name]
	if !ok {
		return nil
	}

	return config
}

// SetDefaultConfigs attempts to run all set-default functions of all registered configs.
func SetDefaultConfigs(v *viper.Viper) {
	mu.RLock()
	defer mu.RUnlock()

	for _, config := range configs {
		config.SetDefault(v)
	}
}

// IterateConfigs iterates through all registered configs.
func IterateConfigs(iterator IteratorFunc) {
	for k, v := range configs {
		if moveNext := iterator(k, v); !moveNext {
			return
		}
	}
}
