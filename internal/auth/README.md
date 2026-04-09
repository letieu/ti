# Auth Module

Authentication module with support for multiple OAuth providers and persistent credential storage.

## Features

- **OAuth 2.0 Support**: Complete OAuth flow implementation
- **Antigravity Provider**: Google Cloud authentication with project discovery
- **Persistent Storage**: Automatic credential persistence using authstore
- **Metadata Support**: Store provider-specific data (email, project_id, etc.)
- **Token Expiry**: Automatic expiry calculation with buffer
- **Multi-provider**: Extensible design for adding more providers

## Structure

```
internal/auth/
├── auth.go           # Core interfaces and types
├── antigravity.go    # Google Cloud (Antigravity) provider
├── storage.go        # Credential persistence helpers
└── server.go         # OAuth callback server
```

## Usage

### Antigravity (Google Cloud) Authentication

```go
import "github.com/letieu/ti/internal/auth"

// Check if already authenticated
if auth.IsAuthenticated(auth.ProviderAntigravity) {
    creds, _ := auth.LoadCredentials(auth.ProviderAntigravity)
    // Use credentials
}

// Perform login
antigravity := auth.AntigravityAuth{}
creds, err := antigravity.Login()
if err != nil {
    log.Fatal(err)
}

// Save credentials for later use
auth.SaveCredentials(creds, auth.ProviderAntigravity)
```

### Working with Metadata

Antigravity provider stores additional metadata:

```go
creds, _ := auth.LoadCredentials(auth.ProviderAntigravity)

// Access metadata
if creds.Metadata != nil {
    email := creds.Metadata["email"]
    projectId := creds.Metadata["project_id"]
}
```

### Adding a New Provider

```go
type MyProvider struct{}

func (p MyProvider) Login() (auth.OAuthCredentials, error) {
    // Implement OAuth flow
    
    return auth.OAuthCredentials{
        Access:  accessToken,
        Refresh: refreshToken,
        Expires: expiresAt,
        Metadata: map[string]string{
            "custom_field": "value",
        },
    }, nil
}

func (p MyProvider) RefreshToken(refreshToken string) auth.OAuthCredentials {
    // Implement token refresh
}
```

## OAuthCredentials Structure

```go
type OAuthCredentials struct {
    Refresh  string            // OAuth refresh token
    Access   string            // OAuth access token
    Expires  int64             // Unix timestamp when token expires
    Metadata map[string]string // Provider-specific metadata
}
```

## Antigravity Provider

The Antigravity provider authenticates with Google Cloud and automatically:

1. **Fetches user email** from Google's userinfo endpoint
2. **Discovers project ID** from Cloud Resource Manager API
3. **Calculates expiry** with 5-minute buffer for safety
4. **Stores metadata** including email and project_id

### Metadata Fields

- `email`: User's Google account email
- `project_id`: First available Google Cloud project ID

### OAuth Scopes

- `cloud-platform`: Full Google Cloud access
- `userinfo.email`: Access to user email
- `userinfo.profile`: Access to user profile
- `cclog`: Cloud logging
- `experimentsandconfigs`: Experiments and configs

## Storage Functions

### SaveCredentials

```go
func SaveCredentials(creds OAuthCredentials, provider string) error
```

Saves OAuth credentials to persistent storage. Metadata is automatically preserved.

### LoadCredentials

```go
func LoadCredentials(provider string) (*OAuthCredentials, error)
```

Loads credentials from storage. Returns error if expired or not found.

### IsAuthenticated

```go
func IsAuthenticated(provider string) bool
```

Checks if valid (non-expired) credentials exist for a provider.

## Provider Constants

```go
const ProviderAntigravity = "antigravity"
```

Use these constants when saving/loading credentials to ensure consistency.

## Example

See `examples/antigravity_auth.go` for a complete working example.

## Security

- OAuth callback server runs on `localhost:51121`
- PKCE (Proof Key for Code Exchange) for enhanced security
- Credentials stored with 0600 permissions via authstore
- 5-minute buffer on token expiry to prevent edge cases

## Testing

The auth module can be tested with:

```bash
go run examples/antigravity_auth.go
```

This will:
1. Open a browser for Google authentication
2. Fetch user email and project ID
3. Save credentials to `~/.config/ti/auth.json`
4. On subsequent runs, load credentials from storage

## Error Handling

Common errors:

- `failed to get user email`: Check OAuth scopes and permissions
- `failed to discover project`: User may not have any GCP projects
- `credentials expired`: Call `Login()` again to refresh
