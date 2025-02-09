package jwt

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type JWTConfig struct {
	SigningKey string // JWT signing key
	// AccessTokenExpiration  time.Duration // Access token duration
	// RefreshTokenExpiration time.Duration // Refresh token duration
	// EnableTokenEndpoint    bool          // Whether to register the /token endpoint
}

// LoadJWTConfig loads the JWT configuration from YAML and environment variables.
func LoadJWTConfig(configName string, configPaths []string) JWTConfig {
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")

	// Set search paths for config files
	for _, path := range configPaths {
		viper.AddConfigPath(path)
	}

	// Set environment variable prefix and enable overrides
	viper.SetEnvPrefix("SECURITY_JWT")
	viper.AutomaticEnv()

	// Define default configuration values
	defaults := map[string]interface{}{
		"security.jwt.signing_key": "default-secret-key",
		// "security.jwt.access_token_expiration":  "15m",
		// "security.jwt.refresh_token_expiration": "168h",
		// "security.jwt.enable_token_endpoint":    true,
	}

	// Apply defaults
	for key, value := range defaults {
		viper.SetDefault(key, value)
	}

	// Read configuration from file
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read JWT config: %v", err)
	}

	// Parse durations with error handling
	// accessTokenDuration := parseDuration(viper.GetString("security.jwt.access_token_expiration"))
	// refreshTokenDuration := parseDuration(viper.GetString("security.jwt.refresh_token_expiration"))

	return JWTConfig{
		SigningKey: viper.GetString("security.jwt.signing_key"),
		// AccessTokenExpiration:  accessTokenDuration,
		// RefreshTokenExpiration: refreshTokenDuration,
		// EnableTokenEndpoint:    viper.GetBool("security.jwt.enable_token_endpoint"),
	}
}

// parseDuration safely parses a duration and logs a fatal error if parsing fails.
func parseDuration(key string) time.Duration {
	duration, err := time.ParseDuration(key)
	if err != nil {
		log.Fatalf("Invalid duration for %s: %v", key, err)
	}
	return duration
}
