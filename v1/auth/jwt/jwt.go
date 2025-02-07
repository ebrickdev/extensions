package jwt

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ebrickdev/ebrick/auth"
	"github.com/ebrickdev/ebrick/transport/httpserver"
)

type StandardClaims = jwt.StandardClaims

// Claims defines custom JWT claims.
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	StandardClaims
}

// JWTAuthenticator implements the Authenticator interface for JWT tokens.
type JWTAuthenticator struct {
	SecretKey  []byte
	Expiration time.Duration
}

func NewJWTAuthenticator() *JWTAuthenticator {
	jwtCfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load JWT config: %v", err)
	}

	return &JWTAuthenticator{
		SecretKey:  []byte(jwtCfg.Secret),
		Expiration: jwtCfg.ParsedExpiration,
	}
}

// Authenticate extracts and validates the JWT token from the HTTP request.
func (j *JWTAuthenticator) Authenticate(r *http.Request) (*auth.User, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("authorization header missing")
	}

	// Expect header in the format "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errors.New("invalid authorization header format")
	}
	tokenString := parts[1]

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return j.SecretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &auth.User{
		ID:       claims.UserID,
		Username: claims.Username,
	}, nil
}

func (j *JWTAuthenticator) Login(c httpserver.Context) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, httpserver.H{"error": "Invalid JSON"})
		return
	}

	// Validate credentials. Replace this dummy check with your real user verification.
	if creds.Username != "admin" || creds.Password != "password" {
		c.JSON(http.StatusUnauthorized, httpserver.H{"error": "Invalid credentials"})
		return
	}

	// Create token claims.
	expirationTime := time.Now().Add(j.Expiration)
	claims := Claims{
		UserID:   "1", // In a real app, this would be the user's ID
		Username: creds.Username,
		StandardClaims: StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Create the JWT token.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.SecretKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.H{"error": "could not generate token"})
		return
	}

	// Respond with the token.
	c.JSON(http.StatusOK, httpserver.H{"token": tokenString})
}
