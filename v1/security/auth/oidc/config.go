package oidc

import (
	"log"

	"github.com/spf13/viper"
)

type OidcConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	IssuerURL    string `mapstructure:"issuer_url"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

func LoadConfig(configName string, configPaths []string) OidcConfig {
	// Set configuration file details
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")

	// Set search paths for config files
	for _, path := range configPaths {
		viper.AddConfigPath(path)
	}

	// Set environment variable prefix and enable overrides
	viper.SetEnvPrefix("SECURITY_OIDC")
	viper.AutomaticEnv()

	viper.SetDefault("security.oidc.client_id", "")
	viper.SetDefault("security.oidc.client_secret", "")
	viper.SetDefault("security.oidc.issuer_url", "")
	viper.SetDefault("security.oidc.redirect_url", "")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read JWT config: %v", err)
	}

	var config OidcConfig
	if err := viper.UnmarshalKey("security.oidc", &config); err != nil {
		log.Fatalf("Failed to unmarshal Oidc config: %v", err)
	}

	log.Default().Println(config)

	return config
}
