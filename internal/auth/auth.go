package auth

import "time"

// OAuthCredentials represents OAuth authentication credentials
// Metadata is optional and can store provider-specific data like email, projectId, etc.
type OAuthCredentials struct {
	Refresh  string            // OAuth refresh token
	Access   string            // OAuth access token
	Expires  int64             // Unix timestamp when token expires
	Metadata map[string]string // Optional provider-specific metadata
}

func (o OAuthCredentials) Expired() bool {
	return time.Now().Unix() >= o.Expires
}

type Auth interface {
	Login() (OAuthCredentials, error)
	RefreshToken(refreshToken string) (OAuthCredentials, error)
}
