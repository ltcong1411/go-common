package registry

import (
	"github.com/spf13/viper"
)

type SetDefaultConfigFunc func(v *viper.Viper)
type GetConfigFunc func(v *viper.Viper) interface{}

// Do nothing set-default function.
var nopSetDefaultFn = SetDefaultConfigFunc(func(v *viper.Viper) {})

// ConfigOpt is a configuration on a config.
type ConfigOpt func(c *configImpl)

//go:generate mockery --name=Config --case=snake
type Config interface {
	SetDefault(v *viper.Viper)
	Get(v *viper.Viper) interface{}
}

// configImpl implements Config.
type configImpl struct {
	setDefaultFn SetDefaultConfigFunc
	getFn        GetConfigFunc
}

// NewConfig returns a new config section with given value getter function.
//
// The value getter function
func NewConfig(getFn GetConfigFunc, opts ...ConfigOpt) Config {
	c := &configImpl{
		setDefaultFn: nopSetDefaultFn,
		getFn:        getFn,
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

// SetDefault set defaults config values to a given Viper object.
func (c *configImpl) SetDefault(v *viper.Viper) {
	c.setDefaultFn(v)
}

// Get retrieves the config values by delegating to its value getter function.
func (c *configImpl) Get(v *viper.Viper) interface{} {
	return c.getFn(v)
}

// WithSetDefault is a ConfigOpt that allows a Config to specify its default values.
func WithSetDefault(setDefaultFn SetDefaultConfigFunc) ConfigOpt {
	return func(c *configImpl) {
		c.setDefaultFn = setDefaultFn
	}
}
