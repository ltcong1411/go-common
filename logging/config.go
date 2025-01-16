package logging

import (
	"github.com/spf13/viper"

	"github.com/ltcong1411/go-common/config"
	"github.com/ltcong1411/go-common/config/registry"
)

// ConfigName is the logging configuration name
const ConfigName = "logging"

// ConfigSectionName is the name of the section used to query configuration information.
//
// To override this name, change it to desired value using logging.ConfigSectionName = "some_name".
var ConfigSectionName = ConfigName

type Config struct {
	IsDevelopment bool
	Level         string
}

func GetConfig(cp config.Provider) *Config {
	return cp.Get(ConfigName).(*Config)
}

func init() {
	registry.RegisterConfig(ConfigName, registry.NewConfig(func(v *viper.Viper) interface{} {
		return &Config{
			IsDevelopment: v.GetBool(ConfigSectionName + ".is_development"),
			Level:         v.GetString(ConfigSectionName + ".level"),
		}
	}, registry.WithSetDefault(func(v *viper.Viper) {
		v.SetDefault(ConfigSectionName, map[string]interface{}{
			"is_development": true,
			"level":          "debug",
		})
	})))
}
