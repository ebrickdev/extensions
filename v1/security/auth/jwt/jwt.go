package jwt

func Init() *JWTAuthManager {
	config := LoadJWTConfig("application", []string{"."})
	authManager := NewJWTAuthManager(config)
	return authManager
}
