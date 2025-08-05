package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Clear any existing environment variables
	os.Clearenv()

	config, err := Load()
	require.NoError(t, err)

	// Test server defaults
	assert.Equal(t, "8080", config.Server.Port)
	assert.Equal(t, 30*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, config.Server.WriteTimeout)
	assert.Equal(t, 60*time.Second, config.Server.IdleTimeout)

	// Test database defaults
	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, "5432", config.Database.Port)
	assert.Equal(t, "messaging_service", config.Database.Name)
	assert.Equal(t, "messaging_user", config.Database.User)
	assert.Equal(t, "messaging_password", config.Database.Password)
	assert.Equal(t, "disable", config.Database.SSLMode)
	assert.Equal(t, 25, config.Database.MaxOpenConns)
	assert.Equal(t, 25, config.Database.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, config.Database.ConnMaxLifetime)
}

func TestLoad_CustomValues(t *testing.T) {
	// Set custom environment variables
	os.Setenv("PORT", "9090")
	os.Setenv("DB_HOST", "custom-host")
	os.Setenv("DB_PASSWORD", "custom-password")
	os.Setenv("SERVER_READ_TIMEOUT", "60s")
	os.Setenv("DB_MAX_OPEN_CONNS", "50")

	config, err := Load()
	require.NoError(t, err)

	// Test custom values
	assert.Equal(t, "9090", config.Server.Port)
	assert.Equal(t, "custom-host", config.Database.Host)
	assert.Equal(t, "custom-password", config.Database.Password)
	assert.Equal(t, 60*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 50, config.Database.MaxOpenConns)

	// Clean up
	os.Clearenv()
}

func TestLoad_InvalidValues(t *testing.T) {
	// Test invalid duration - should fall back to default
	os.Setenv("SERVER_READ_TIMEOUT", "invalid")
	config, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, config.Server.ReadTimeout) // Should use default

	// Test invalid integer - should fall back to default
	os.Clearenv()
	os.Setenv("DB_MAX_OPEN_CONNS", "not-a-number")
	config, err = Load()
	require.NoError(t, err)
	assert.Equal(t, 25, config.Database.MaxOpenConns) // Should use default

	// Clean up
	os.Clearenv()
}

func TestDatabaseConfig_GetDSN(t *testing.T) {
	dbConfig := DatabaseConfig{
		Host:     "test-host",
		Port:     "5433",
		Name:     "test-db",
		User:     "test-user",
		Password: "test-pass",
		SSLMode:  "require",
	}

	dsn := dbConfig.GetDSN()
	expected := "host=test-host port=5433 dbname=test-db user=test-user password=test-pass sslmode=require"
	assert.Equal(t, expected, dsn)
}

func TestConfig_Validate(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Port:         "8080",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            "5432",
			Name:            "test",
			User:            "user",
			Password:        "pass",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    25,
			ConnMaxLifetime: 5 * time.Minute,
		},
	}

	err := config.validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_Errors(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "empty server port",
			config: &Config{
				Server: ServerConfig{Port: ""},
			},
			expectError: true,
		},
		{
			name: "empty database host",
			config: &Config{
				Server: ServerConfig{Port: "8080"},
				Database: DatabaseConfig{
					Host: "",
					Name: "test",
					User: "user",
				},
			},
			expectError: true,
		},
		{
			name: "empty database name",
			config: &Config{
				Server: ServerConfig{Port: "8080"},
				Database: DatabaseConfig{
					Host: "localhost",
					Name: "",
					User: "user",
				},
			},
			expectError: true,
		},
		{
			name: "empty database user",
			config: &Config{
				Server: ServerConfig{Port: "8080"},
				Database: DatabaseConfig{
					Host: "localhost",
					Name: "test",
					User: "",
				},
			},
			expectError: true,
		},
		{
			name: "invalid read timeout",
			config: &Config{
				Server: ServerConfig{
					Port:        "8080",
					ReadTimeout: -1 * time.Second,
				},
				Database: DatabaseConfig{
					Host: "localhost",
					Name: "test",
					User: "user",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
