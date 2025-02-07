package jwt

import (
	"fmt"
	"time"

	"github.com/ebrickdev/ebrick/config"
)

// Config holds the top-level configuration loaded from the YAML file.
type Config struct {
	Auth AuthConfig `yaml:"auth"`
}

// AuthConfig holds authentication-related configuration.
type AuthConfig struct {
	Jwt JwtConfig `yaml:"jwt"`
}

// JwtConfig holds the JWT-specific configuration.
type JwtConfig struct {
	Secret           string        `yaml:"secret"`
	Expiration       string        `yaml:"expiration"`
	ParsedExpiration time.Duration `yaml:"-"`
}

// LoadConfig loads the JWT configuration from the application.yaml file.
// It returns a pointer to JwtConfig or an error if loading or parsing fails.
func LoadConfig() (*JwtConfig, error) {
	var cfg Config
	if err := config.LoadConfig("application", []string{"."}, &cfg); err != nil {
		return nil, fmt.Errorf("JWT: error loading config: %w", err)
	}

	// If an expiration duration is specified, parse it.
	if cfg.Auth.Jwt.Expiration != "" {
		expiration, err := time.ParseDuration(cfg.Auth.Jwt.Expiration)
		if err != nil {
			return nil, fmt.Errorf("JWT: error parsing expiration config: %w", err)
		}
		cfg.Auth.Jwt.ParsedExpiration = expiration
	}

	return &cfg.Auth.Jwt, nil
}
