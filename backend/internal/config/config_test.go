package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestIsDev_True(t *testing.T) {
	cfg := &Config{Env: "development"}
	assert.True(t, cfg.IsDev())
}

func TestIsDev_False(t *testing.T) {
	cfg := &Config{Env: "production"}
	assert.False(t, cfg.IsDev())
}

func TestLoad_Development(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	os.Setenv("ENV", "development")
	os.Setenv("PORT", "8080")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("JWT_ACCESS_SECRET", "test-access")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh")

	t.Cleanup(func() {
		os.Unsetenv("ENV")
		os.Unsetenv("PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("JWT_ACCESS_SECRET")
		os.Unsetenv("JWT_REFRESH_SECRET")
	})

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "development", cfg.Env)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "5432", cfg.DBPort)
	assert.Equal(t, "testdb", cfg.DBName)
	assert.Equal(t, "testuser", cfg.DBUser)
	assert.Equal(t, "testpass", cfg.DBPassword)
	assert.Equal(t, "test-access", cfg.JWTAccessSecret)
	assert.Equal(t, "test-refresh", cfg.JWTRefreshSecret)

	assert.Equal(t, "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable", cfg.DBURL)

	assert.Equal(t, 31337, cfg.ChainID)
	assert.Equal(t, "Hardhat", cfg.ChainName)
}

func TestLoad_DefaultEnv(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	os.Setenv("JWT_ACCESS_SECRET", "test-access")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh")

	t.Cleanup(func() {
		os.Unsetenv("JWT_ACCESS_SECRET")
		os.Unsetenv("JWT_REFRESH_SECRET")
	})

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "development", cfg.Env)
	assert.True(t, cfg.IsDev())
}
