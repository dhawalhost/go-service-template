package config

import kitconfig "github.com/dhawalhost/gokit/config"

// Config embeds gokit base config and adds service-specific fields.
// RENAME_ME: Add your own service-specific config sections here.
type Config struct {
	kitconfig.Config `mapstructure:",squash"`
}

// Load reads config from environment variables and/or a YAML file.
func Load() (*Config, error) {
	base, err := kitconfig.Load("")
	if err != nil {
		return nil, err
	}
	return &Config{Config: *base}, nil
}

// MustLoad is a helper that panics on config load error, for use in main() where we want to fail fast.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}
