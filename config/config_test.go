package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_LoadsCORSOriginsFromYAML(t *testing.T) {
	t.Setenv("APP_CONFIG_FILE", writeTempConfig(t, `
server:
  addr: ":8080"
database:
  dsn: postgres://pguser:password@localhost:5432/service_db?sslmode=disable
redis:
  addr: localhost:6379
cors:
  allowed_origins:
    - https://a.example.com
    - https://b.example.com
`))
	unsetEnv(t, "APP_CORS_ALLOWED_ORIGINS")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"https://a.example.com", "https://b.example.com"}, cfg.CORS.AllowedOrigins)
}

func TestLoad_CORSOriginsEnvOverridesYAML(t *testing.T) {
	t.Setenv("APP_CONFIG_FILE", writeTempConfig(t, `
server:
  addr: ":8080"
database:
  dsn: postgres://pguser:password@localhost:5432/service_db?sslmode=disable
redis:
  addr: localhost:6379
cors:
  allowed_origins:
    - https://yaml.example.com
`))
	t.Setenv("APP_CORS_ALLOWED_ORIGINS", " https://env-1.example.com, , https://env-2.example.com ")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"https://env-1.example.com", "https://env-2.example.com"}, cfg.CORS.AllowedOrigins)
}

func TestLoad_TenancyLoadedFromYAML(t *testing.T) {
	t.Setenv("APP_CONFIG_FILE", writeTempConfig(t, `
server:
  addr: ":8080"
database:
  dsn: postgres://pguser:password@localhost:5432/service_db?sslmode=disable
redis:
  addr: localhost:6379
tenancy:
  enabled: true
  default_tenant_id: team-a
`))

	cfg, err := Load()
	require.NoError(t, err)
	assert.True(t, cfg.Tenancy.Enabled)
	assert.Equal(t, "team-a", cfg.Tenancy.DefaultTenantID)
}

func TestLoad_TenancyEnvOverridesYAML(t *testing.T) {
	t.Setenv("APP_CONFIG_FILE", writeTempConfig(t, `
server:
  addr: ":8080"
database:
  dsn: postgres://pguser:password@localhost:5432/service_db?sslmode=disable
redis:
  addr: localhost:6379
tenancy:
  enabled: false
  default_tenant_id: yaml-tenant
`))
	t.Setenv("APP_TENANCY_ENABLED", "true")
	t.Setenv("APP_TENANCY_DEFAULT_TENANT_ID", "env-tenant")

	cfg, err := Load()
	require.NoError(t, err)
	assert.True(t, cfg.Tenancy.Enabled)
	assert.Equal(t, "env-tenant", cfg.Tenancy.DefaultTenantID)
}

func TestLoad_TenancyEnabledDefaultsTenantID(t *testing.T) {
	t.Setenv("APP_CONFIG_FILE", writeTempConfig(t, `
server:
  addr: ":8080"
database:
  dsn: postgres://pguser:password@localhost:5432/service_db?sslmode=disable
redis:
  addr: localhost:6379
`))
	t.Setenv("APP_TENANCY_ENABLED", "true")
	unsetEnv(t, "APP_TENANCY_DEFAULT_TENANT_ID")

	cfg, err := Load()
	require.NoError(t, err)
	assert.True(t, cfg.Tenancy.Enabled)
	assert.Equal(t, "default", cfg.Tenancy.DefaultTenantID)
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	oldValue, existed := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))

	t.Cleanup(func() {
		if existed {
			require.NoError(t, os.Setenv(key, oldValue))
			return
		}
		require.NoError(t, os.Unsetenv(key))
	})
}
