package nats

import (
	"log"

	"github.com/ebrickdev/ebrick/config"
	"github.com/ebrickdev/ebrick/event"
)

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

func init() {
	// Get the database configuration from the config package
	var cfg Config
	err := config.LoadConfig("application", []string{"."}, &cfg)
	if err != nil {
		log.Fatalf("Nats: error loading config %v", err)
	}
	// Initialize NATS connection
	log.Printf("Nats: Connecting to nats on %s \n", cfg.Messaging.Nats.URL)
	eventBus, err := NewEventBus(cfg.Messaging.Nats.URL, cfg.Messaging.Nats.Username, cfg.Messaging.Nats.Password)
	if err != nil {
		log.Fatalf("Nats: error initializing event bus. %v", err)
	}
	event.DefaultEventBus = eventBus
	log.Printf("Nats: Connected to Nats on %s \n", cfg.Messaging.Nats.URL)
}
