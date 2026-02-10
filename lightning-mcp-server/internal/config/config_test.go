// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test LoadConfig with default values.
func TestLoadConfig_Defaults(t *testing.T) {
	// Clear environment variables to test defaults.
	os.Unsetenv("DEVELOPMENT")
	os.Unsetenv("LNC_DEFAULT_MAILBOX")
	os.Unsetenv("LNC_DEFAULT_TIMEOUT")

	config := LoadConfig()

	assert.Equal(t, "lightning-mcp-server", config.ServerName)
	assert.Equal(t, "1.0.0", config.ServerVersion)
	assert.False(t, config.Development) // Default is false.
	assert.Equal(t, "mailbox.terminal.lightning.today:443",
		config.DefaultMailboxServer)
	assert.Equal(t, 30*time.Second, config.DefaultTimeout)
	assert.Equal(t, 30*time.Second, config.ShutdownTimeout)
}

// Test LoadConfig with environment variables.
func TestLoadConfig_EnvironmentVariables(t *testing.T) {
	// Set environment variables.
	os.Setenv("DEVELOPMENT", "false")
	os.Setenv("LNC_DEFAULT_MAILBOX", "test.server:443")
	os.Setenv("LNC_DEFAULT_TIMEOUT", "45s")

	defer func() {
		os.Unsetenv("DEVELOPMENT")
		os.Unsetenv("LNC_DEFAULT_MAILBOX")
		os.Unsetenv("LNC_DEFAULT_TIMEOUT")
	}()

	config := LoadConfig()

	assert.Equal(t, "lightning-mcp-server", config.ServerName)
	assert.Equal(t, "1.0.0", config.ServerVersion)
	assert.False(t, config.Development)
	assert.Equal(t, "test.server:443", config.DefaultMailboxServer)
	assert.Equal(t, 45*time.Second, config.DefaultTimeout)
}

// Test environment variable parsing for development mode.
func TestLoadConfig_DevelopmentMode(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "true",
			envValue: "true",
			expected: true,
		},
		{
			name:     "false",
			envValue: "false",
			expected: false,
		},
		{
			name:     "1",
			envValue: "1",
			expected: true,
		},
		{
			name:     "0",
			envValue: "0",
			expected: false,
		},
		{
			name:     "yes",
			envValue: "yes",
			expected: false, // Default fallback when parse fails.
		},
		{
			name:     "invalid",
			envValue: "invalid",
			expected: false, // Default fallback when parse fails.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("DEVELOPMENT", tt.envValue)
			defer os.Unsetenv("DEVELOPMENT")

			config := LoadConfig()
			assert.Equal(t, tt.expected, config.Development)
		})
	}
}

// Test config validation.
func TestConfig_Validation(t *testing.T) {
	config := LoadConfig()

	// Test that all required fields are set.
	assert.NotEmpty(t, config.ServerName)
	assert.NotEmpty(t, config.ServerVersion)
	assert.NotEmpty(t, config.DefaultMailboxServer)
	assert.NotZero(t, config.DefaultTimeout)
	assert.NotZero(t, config.ShutdownTimeout)
	assert.NotZero(t, config.ConnectionTimeout)
	assert.NotZero(t, config.MaxConnectionRetries)
}

// Test config structure completeness.
func TestConfig_Structure(t *testing.T) {
	config := LoadConfig()

	// Verify all fields are accessible.
	_ = config.ServerName
	_ = config.ServerVersion
	_ = config.Development
	_ = config.DefaultMailboxServer
	_ = config.DefaultTimeout
	_ = config.DefaultDevMode
	_ = config.DefaultInsecure
	_ = config.MaxConnectionRetries
	_ = config.ConnectionTimeout
	_ = config.ShutdownTimeout

	// Test that we can modify fields.
	config.ServerName = "modified"
	assert.Equal(t, "modified", config.ServerName)
}

// Test getEnvString helper function.
func TestGetEnvString(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		expected     string
	}{
		{
			name:         "environment_variable_set",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "from_env",
			setEnv:       true,
			expected:     "from_env",
		},
		{
			name:         "environment_variable_not_set",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "",
			setEnv:       false,
			expected:     "default",
		},
		{
			name:         "empty_environment_variable",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "",
			setEnv:       true,
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvString(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test getEnvBool helper function.
func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		setEnv       bool
		expected     bool
	}{
		{
			name:         "true_value",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "true",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "false_value",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "false",
			setEnv:       true,
			expected:     false,
		},
		{
			name:         "invalid_value_uses_default",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "invalid",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "not_set_uses_default",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "",
			setEnv:       false,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvBool(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test timeout values are reasonable.
func TestConfig_TimeoutValues(t *testing.T) {
	config := LoadConfig()

	// Verify timeout values are reasonable.
	assert.True(t, config.DefaultTimeout > 0)
	assert.True(t, config.ShutdownTimeout > 0)
	assert.True(t, config.ConnectionTimeout > 0)
	assert.True(t, config.DefaultTimeout >= 5*time.Second,
		"Default timeout should be at least 5 seconds")
	assert.True(t, config.ShutdownTimeout >= 5*time.Second,
		"Shutdown timeout should be at least 5 seconds")
	assert.True(t, config.ConnectionTimeout >= 5*time.Second,
		"Connection timeout should be at least 5 seconds")
}

// Benchmark config creation.
func BenchmarkLoadConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = LoadConfig()
	}
}

// Benchmark environment variable lookup.
func BenchmarkGetEnvString(b *testing.B) {
	os.Setenv("BENCH_TEST", "value")
	defer os.Unsetenv("BENCH_TEST")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getEnvString("BENCH_TEST", "default")
	}
}
