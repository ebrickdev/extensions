package go_cache

import (
	"log"
	"time"

	"github.com/ebrickdev/ebrick/cache"
	"github.com/ebrickdev/ebrick/cache/store"
	"github.com/ebrickdev/ebrick/config"
	gocache "github.com/patrickmn/go-cache"
)

type Config struct {
	Cache CacheConfig `yaml:"cache"`
}

type CacheConfig struct {
	Expiration      int `yaml:"expiration"`
	CleanupInterval int `yaml:"cleanup_interval"`
}

func init() {
	// Get the database configuration from the config package
	var cfg Config
	err := config.LoadConfig("application", []string{"."}, &cfg)
	if err != nil {
		log.Fatalf("gocache: error loading config %v, using default cache configuration", err)
	}

	// Set default values if not set
	if cfg.Cache.Expiration == 0 {
		cfg.Cache.Expiration = 300 // Default to 5 minutes (300 seconds)
		log.Println("gocache: Expiration not set, using default value of 5 minutes")
	}
	if cfg.Cache.CleanupInterval == 0 {
		cfg.Cache.CleanupInterval = 600 // Default to 10 minutes (600 seconds)
		log.Println("gocache: CleanupInterval not set, using default value of 10 minutes")
	}

	// Initialize the GoCache with the loaded or default configuration
	c := gocache.New(
		time.Duration(cfg.Cache.Expiration)*time.Second,
		time.Duration(cfg.Cache.CleanupInterval)*time.Second,
	)
	gcstore := NewGoCache(c, store.WithExpiration(time.Duration(cfg.Cache.Expiration)))
	cache.DefaultCache = cache.New(gcstore)
	log.Println("gocache: GoCache Initialized")
}
