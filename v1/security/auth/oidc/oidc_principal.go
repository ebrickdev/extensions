package oidc

// OidcPrincipal represents an authenticated user's principal.
type OidcPrincipal struct {
	id     string
	email  string
	roles  []string
	claims map[string]interface{}
}

// GetID returns the principal's unique identifier.
func (p *OidcPrincipal) GetID() string {
	return p.id
}

// GetEmail returns the principal's email address.
func (p *OidcPrincipal) GetEmail() string {
	return p.email
}

// GetRoles returns the roles assigned to the principal.
func (p *OidcPrincipal) GetRoles() []string {
	return p.roles
}

// GetClaims returns all claims associated with the principal.
func (p *OidcPrincipal) GetClaims() map[string]interface{} {
	return p.claims
}
