package inmemory

type Config struct {
	Messaging MessagingConfig `yaml:"messaging"`
}

type MessagingConfig struct {
	Memory MemoryConfig `yaml:"memory"`
}
type MemoryConfig struct {
	// Add fields here
}
