package config

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/spf13/viper"

	"github.com/fatih/structs"
	"github.com/iancoleman/strcase"
	"github.com/ltcong1411/go-common/config/registry"
)

const (
	MaskedTagOption = "masked"
	MaskedValue     = "******"
)

var (
	defaultMaskedFieldNames = map[string]bool{
		"password": true,
		"key":      true,
		"secret":   true,
	}
)

// Provider represents the configuration provider.
//
//go:generate mockery --name=Provider --case=snake
type Provider interface {
	// Get returns a configuration with given name.
	Get(name string) interface{}

	// DumpConfigs compiles all registered config values into a map.
	DumpConfigs() map[string]interface{}
}

// ProviderOpt is an option on a given Provider.
type ProviderOpt func(impl *providerImpl)

// ViperInitializer is a function that initializes a given Viper object.
type ViperInitializer func(v *viper.Viper)

// providerImpl implements Provider interface.
type providerImpl struct {
	viper            *viper.Viper
	viperInitializer ViperInitializer

	once *sync.Once
}

// NewProvider creates a new instance of configuration provider.
func NewProvider(opts ...ProviderOpt) Provider {
	provider := &providerImpl{
		viper: viper.New(),
		viperInitializer: func(v *viper.Viper) {
			v.SetConfigName("app-config") // name of config file (without extension)
			v.SetConfigType("yaml")
			v.AddConfigPath("$APP_CONFIG_DIR")
			v.AddConfigPath(".")
			v.AddConfigPath("$HOME")
			v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
			v.SetEnvPrefix("APP")
			v.AutomaticEnv()
		},
		once: &sync.Once{},
	}

	for _, o := range opts {
		o(provider)
	}

	return provider
}

func (impl *providerImpl) initializeOnce() {
	impl.once.Do(func() {
		registry.SetDefaultConfigs(impl.viper)
		impl.viperInitializer(impl.viper)
		_ = impl.viper.ReadInConfig()
	})
}

func (impl *providerImpl) Get(name string) interface{} {
	impl.initializeOnce()

	config := registry.GetConfig(name)
	if config == nil {
		panic(fmt.Sprintf("cannot find configuration with name %s", name))
	}

	return config.Get(impl.viper)
}

// DumpConfigs compiles all registered config values into a map.
func (impl *providerImpl) DumpConfigs() map[string]interface{} {
	impl.initializeOnce()

	values := map[string]interface{}{}

	registry.IterateConfigs(func(name string, config registry.Config) bool {
		values[name] = impl.formatConfig(config.Get(impl.viper))
		return true
	})

	return values
}

// formatConfig ensures the config value is formatted and masked.
func (impl *providerImpl) formatConfig(value interface{}) interface{} {
	if !structs.IsStruct(value) {
		return value
	}

	v := structs.New(value)
	fields := v.Fields()

	for _, f := range fields {
		impl.formatConfigField(f)
	}

	return v.Map()
}

func (impl *providerImpl) formatConfigField(f *structs.Field) {
	if (f.Kind() == reflect.Struct || f.Kind() == reflect.Ptr) && f.Fields() != nil {
		for _, childF := range f.Fields() {
			impl.formatConfigField(childF)
		}
	}

	tagName, tagOpts := parseTag(f.Tag(structs.DefaultTagName))

	if tagOpts.Has(MaskedTagOption) || isMaskedFieldNameOrTag(tagName) || isMaskedFieldNameOrTag(f.Name()) {
		if f.Kind() == reflect.String {
			_ = f.Set(MaskedValue)
		} else {
			panic(fmt.Errorf("masked field '%s' must be string", f.Name()))
		}
	}
}

func isMaskedFieldNameOrTag(name string) bool {
	formatted := strcase.ToSnake(name)
	parts := strings.Split(formatted, "_")

	for _, v := range parts {
		_, contained := defaultMaskedFieldNames[v]
		if contained {
			return true
		}
	}

	return false
}

// WithViperInitializer returns an option that allows setting of viper initializer in the provider.
func WithViperInitializer(viperInitializer ViperInitializer) ProviderOpt {
	return func(impl *providerImpl) {
		impl.viperInitializer = viperInitializer
	}
}
