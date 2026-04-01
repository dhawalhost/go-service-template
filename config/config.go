package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	kitconfig "github.com/dhawalhost/gokit/config"
	"gopkg.in/yaml.v3"
)

// Config embeds gokit base config and adds service-specific fields.
// RENAME_ME: Add your own service-specific config sections here.
type Config struct {
	kitconfig.Config `mapstructure:",squash"`
	CORS             CORSConfig    `mapstructure:"cors"`
	Tenancy          TenancyConfig `mapstructure:"tenancy"`
}

// CORSConfig holds CORS middleware configuration.
type CORSConfig struct {
	// AllowedOrigins is a list of allowed origin domains for CORS requests.
	// Example: ["https://example.com", "https://app.example.com"]
	// Leave empty to restrict all CORS requests.
	AllowedOrigins []string `mapstructure:"allowed_origins" yaml:"allowed_origins"`
}

// TenancyConfig controls optional tenant scoping for request handling and data access.
type TenancyConfig struct {
	// Enabled turns on tenant extraction and tenant-scoped repository filters.
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`
	// DefaultTenantID is used when tenancy is enabled but the request carries no tenant ID.
	DefaultTenantID string `mapstructure:"default_tenant_id" yaml:"default_tenant_id"`
}

// Load reads config from a YAML file (APP_CONFIG_FILE env var) and/or
// APP_* environment variables.
//
// When APP_CONFIG_FILE is set, Viper loads the YAML first which makes all
// keys known, so Unmarshal correctly populates every nested struct field.
// When no YAML file is provided, critical fields (DSN, Redis addr) are read
// directly via os.Getenv as a fallback for the known Viper AutomaticEnv +
// Unmarshal limitation with nested structs that have no default set.
func Load() (*Config, error) {
	cfgFile := os.Getenv("APP_CONFIG_FILE")
	base, err := kitconfig.Load(cfgFile)
	if err != nil {
		return nil, err
	}
	cfg := &Config{Config: *base}

	if strings.TrimSpace(cfgFile) != "" {
		overrides, err := loadOverridesFromYAML(cfgFile)
		if err != nil {
			return nil, err
		}
		cfg.CORS = overrides.CORS
		cfg.Tenancy = overrides.Tenancy
	}

	if raw, ok := os.LookupEnv("APP_CORS_ALLOWED_ORIGINS"); ok {
		cfg.CORS.AllowedOrigins = parseCSV(raw)
	}
	if raw, ok := os.LookupEnv("APP_TENANCY_ENABLED"); ok {
		enabled, err := strconv.ParseBool(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("config: parse APP_TENANCY_ENABLED: %w", err)
		}
		cfg.Tenancy.Enabled = enabled
	}
	if raw, ok := os.LookupEnv("APP_TENANCY_DEFAULT_TENANT_ID"); ok {
		cfg.Tenancy.DefaultTenantID = strings.TrimSpace(raw)
	}

	// Fallback for env-only runs where Viper Unmarshal misses undefaulted keys.
	if cfg.Database.DSN == "" {
		cfg.Database.DSN = os.Getenv("APP_DATABASE_DSN")
	}
	if cfg.Redis.Addr == "" {
		cfg.Redis.Addr = os.Getenv("APP_REDIS_ADDR")
	}
	if strings.TrimSpace(cfg.Database.DSN) == "" {
		return nil, fmt.Errorf("config: APP_DATABASE_DSN is required; set APP_CONFIG_FILE or export APP_DATABASE_DSN")
	}
	cfg.Tenancy = normalizeTenancy(cfg.Tenancy)
	return cfg, nil
}

type yamlConfig struct {
	CORS    CORSConfig    `yaml:"cors"`
	Tenancy TenancyConfig `yaml:"tenancy"`
}

func loadOverridesFromYAML(path string) (*yamlConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read APP_CONFIG_FILE: %w", err)
	}

	var cfg yamlConfig
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse APP_CONFIG_FILE: %w", err)
	}

	return &cfg, nil
}

func parseCSV(value string) []string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		origins = append(origins, part)
	}
	return origins
}

func normalizeTenancy(cfg TenancyConfig) TenancyConfig {
	if cfg.Enabled && strings.TrimSpace(cfg.DefaultTenantID) == "" {
		cfg.DefaultTenantID = "default"
	}
	return cfg
}

// MustLoad is a helper that panics on config load error, for use in main() where we want to fail fast.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}
