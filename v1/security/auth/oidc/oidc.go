package oidc

import (
	"context"
)

func Init() *OIDCAuthManager {
	config := LoadConfig("application", []string{"."})
	authManager := NewOIDCAuthManager(context.Background(), config)
	return authManager
}
