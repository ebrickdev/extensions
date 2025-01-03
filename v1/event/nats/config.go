package nats

type Config struct {
	Messaging MessagingConfig `yaml:"messaging"`
}

type MessagingConfig struct {
	Nats NatsConfig `yaml:"nats"`
}

type NatsConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
