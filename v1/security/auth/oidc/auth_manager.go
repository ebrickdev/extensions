package oidc

import (
	"context"
	"fmt"
	"log"

	"github.com/coreos/go-oidc"
	"github.com/ebrickdev/ebrick/security/auth"
	"golang.org/x/oauth2"
)

// OIDCAuthManager handles authentication using an OIDC provider.
type OIDCAuthManager struct {
	provider     *oidc.Provider
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
}

// NewOIDCAuthManager initializes and returns a new OIDCAuthManager.
// The provided context is used for the provider initialization.
// It returns an error rather than exiting the application.
func NewOIDCAuthManager(ctx context.Context, config OidcConfig) *OIDCAuthManager {
	// Optional: log the issuer URL for debugging purposes.
	log.Printf("Initializing OIDC provider with issuer URL: %s", config.IssuerURL)

	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		log.Fatalf("failed to get provider: %v", err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: config.ClientID})

	return &OIDCAuthManager{
		provider:     provider,
		oauth2Config: oauth2Config,
		verifier:     verifier,
	}
}

// AuthCodeURL returns the URL to which users should be redirected for authentication.
// The state parameter is used to help mitigate CSRF attacks.
func (m *OIDCAuthManager) AuthCodeURL(state string) string {
	return m.oauth2Config.AuthCodeURL(state)
}

// Exchange exchanges an authorization code for an OAuth2 token.
func (m *OIDCAuthManager) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return m.oauth2Config.Exchange(ctx, code)
}

// VerifyIDToken verifies the provided raw ID token string and returns the parsed token.
func (m *OIDCAuthManager) VerifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return m.verifier.Verify(ctx, rawIDToken)
}

// GetUserInfo retrieves the user information from the OIDC provider using the OAuth2 token.
func (m *OIDCAuthManager) GetUserInfo(ctx context.Context, token *oauth2.Token) (*oidc.UserInfo, error) {
	return m.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
}

// Authenticate verifies the token, fetches user info, and returns an OidcPrincipal.
// It decodes the ID token claims and, if present, extracts the roles.
func (m *OIDCAuthManager) Authenticate(ctx context.Context, token string) (auth.Principal, error) {
	// Verify the ID token.
	idToken, err := m.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Retrieve user info.
	userInfo, err := m.GetUserInfo(ctx, &oauth2.Token{AccessToken: token})
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Decode ID token claims.
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse ID token claims: %w", err)
	}

	// Attempt to extract roles from claims (if available).
	var roles []string
	if r, ok := claims["roles"]; ok {
		if rolesInterface, ok := r.([]interface{}); ok {
			for _, roleInterface := range rolesInterface {
				if roleStr, ok := roleInterface.(string); ok {
					roles = append(roles, roleStr)
				}
			}
		}
	}

	return &OidcPrincipal{
		id:     idToken.Subject,
		email:  userInfo.Email,
		roles:  roles,
		claims: claims,
	}, nil
}
