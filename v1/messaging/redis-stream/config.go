package redisstream

type RedisStreamConfig struct {
	URL               string `mapstructure:"url"`
	Username          string `mapstructure:"username"`
	Password          string `mapstructure:"password"`
	TLS               bool   `mapstructure:"tls"`               // Secure connection flag
	MaxRetries        int    `mapstructure:"maxRetries"`        // For retry logic
	MaxConnections    int    `mapstructure:"maxConnections"`    // Connection pooling config
	ConnectionTimeout int    `mapstructure:"connectionTimeout"` // Timeout for connections
}
