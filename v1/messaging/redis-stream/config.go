package redisstream

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type RedisStreamConfig struct {
	URL               string `mapstructure:"url"`
	Username          string `mapstructure:"username"`
	Password          string `mapstructure:"password"`
	TLS               bool   `mapstructure:"tls"`                // Secure connection flag
	MaxRetries        int    `mapstructure:"max_retries"`        // For retry logic
	MaxConnections    int    `mapstructure:"max_connections"`    // Connection pooling config
	ConnectionTimeout int    `mapstructure:"connection_timeout"` // Timeout for connections
}

func LoadConfig(configName string, configPaths []string) RedisStreamConfig {
	// Set configuration file details
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")

	// Set search paths for config files
	for _, path := range configPaths {
		viper.AddConfigPath(path)
	}

	viper.SetDefault("messaging.redis.url", "localhost:6379")
	// Set environment variable prefix and enable overrides

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	var config RedisStreamConfig
	if err := viper.UnmarshalKey("messaging.redis", &config); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	return config
}
