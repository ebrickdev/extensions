package jwt

// JWTPrincipal implements the Principal interface.
type JWTPrincipal struct {
	id     string
	email  string
	roles  []string
	claims map[string]interface{}
}

func (p *JWTPrincipal) GetID() string {
	return p.id
}

func (p *JWTPrincipal) GetEmail() string {
	return p.email
}

func (p *JWTPrincipal) GetRoles() []string {
	return p.roles
}

func (p *JWTPrincipal) GetClaims() map[string]interface{} {
	return p.claims
}
