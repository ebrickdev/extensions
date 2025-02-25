package redis

import (
	"fmt"
	"log"
	"time"

	"github.com/ebrickdev/ebrick/cache"
	"github.com/ebrickdev/ebrick/cache/store"
	"github.com/ebrickdev/ebrick/config"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Cache CacheConfig `yaml:"cache"`
}

type CacheConfig struct {
	Expiration           int         `yaml:"default_expiration"`
	ClientSideExpiration int         `yaml:"client_side_expiration"`
	CleanupInterval      int         `yaml:"cleanup_interval"`
	Redis                RedisConfig `yaml:"redis"`
}

type RedisConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

func Init() cache.Cache {
	// Get the database configuration from the config package
	var cfg Config
	err := config.LoadConfig("application", []string{"."}, &cfg)
	if err != nil {
		log.Fatalf("Redis: error loading config %v", err)
	}
	cli := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Cache.Redis.Host, cfg.Cache.Redis.Port),
		Username: cfg.Cache.Redis.Username,
		Password: cfg.Cache.Redis.Password,
		DB:       0,
	})
	// Set default values if not set
	if cfg.Cache.Expiration == 0 {
		cfg.Cache.Expiration = 300 // Default to 5 minutes (300 seconds)
		log.Println("Redis: Expiration not set, using default value of 5 minutes")
	}

	if cfg.Cache.ClientSideExpiration == 0 {
		cfg.Cache.ClientSideExpiration = 300 // Default to 5 minutes (300 seconds)
		log.Println("Redis: ClientSideExpiration not set, using default value of 5 minutes")
	}

	return cache.New(NewRedis(
		cli,
		store.WithExpiration(time.Duration(cfg.Cache.Expiration)),
		store.WithClientSideCaching(time.Duration(cfg.Cache.ClientSideExpiration))))

}
