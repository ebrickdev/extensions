package postgresql

import (
	"fmt"
	"log"

	"github.com/ebrickdev/ebrick/config"
	"github.com/ebrickdev/ebrick/db"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
}
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	Enable   bool
	Type     string
}

// InitDB initializes the PostgreSQL database connection and returns a *gorm.DB instance.
func init() {
	// Get the database configuration from the config package
	var cfg Config
	err := config.LoadConfig("application", []string{"."}, &cfg)
	if err != nil {
		log.Fatalf("PostgreSQL: error loading database config %v", err)
	}
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=%s password=%s",
		cfg.Database.Host, cfg.Database.User, cfg.Database.DBName, cfg.Database.SSLMode, cfg.Database.Password)

	// Open a connection to the database
	ds, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("PostgreSQL: failed to connect to database %v", err)
	}
	db.DefaultDataSource = ds
	log.Println("PostgreSQL: Connected to PostgreSQL database")
}
