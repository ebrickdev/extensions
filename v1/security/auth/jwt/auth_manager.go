package jwt

import (
	"context"
	"errors"

	"github.com/ebrickdev/ebrick/security/auth"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuthManager implements local authentication using JWT.
type JWTAuthManager struct {
	config JWTConfig
}

func NewJWTAuthManager(config JWTConfig) *JWTAuthManager {
	return &JWTAuthManager{
		config: config,
	}
}

// Authenticate validates a JWT token and returns the corresponding Principal.
func (j *JWTAuthManager) Authenticate(ctx context.Context, token string) (auth.Principal, error) {
	claims, err := validateJWT(token, j.config.SigningKey)
	if err != nil {
		return nil, err
	}
	return &JWTPrincipal{
		id:     claims["sub"].(string),
		email:  claims["email"].(string),
		roles:  parseRoles(claims),
		claims: claims,
	}, nil
}

func parseRoles(claims map[string]interface{}) []string {
	if roles, ok := claims["roles"].([]interface{}); ok {
		var result []string
		for _, role := range roles {
			result = append(result, role.(string))
		}
		return result
	}
	return nil
}

// validateJWT parses and validates the JWT token.
func validateJWT(tokenString string, signingKey string) (map[string]interface{}, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(signingKey), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}
	return nil, errors.New("failed to parse token claims")
}
