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
