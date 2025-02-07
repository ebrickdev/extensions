package redisstream

type Config struct {
	Messaging MessagingConfig `yaml:"messaging"`
}

type MessagingConfig struct {
	RedisStream RedisStreamConfig `yaml:"redis"`
}

type RedisStreamConfig struct {
	URL               string `yaml:"url"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	TLS               bool   `yaml:"tls"`                // Secure connection flag
	MaxRetries        int    `yaml:"max_retries"`        // For retry logic
	MaxConnections    int    `yaml:"max_connections"`    // Connection pooling config
	ConnectionTimeout int    `yaml:"connection_timeout"` // Timeout for connections
}
