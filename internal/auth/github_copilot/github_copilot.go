package githubcopilot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/letieu/ti/internal/auth"
)

const (
	clientID = "Iv1.b507a08c87ecfe98"

	deviceCodeURL   = "https://github.com/login/device/code"
	accessTokenURL  = "https://github.com/login/oauth/access_token"
	copilotTokenURL = "https://api.github.com/copilot_internal/v2/token"

	userAgent = "GitHubCopilotChat/0.32.4"
)

var ErrNotAvailable = errors.New("github copilot not available")

// GitHubCopilotAuth implements the auth.Auth interface
type GitHubCopilotAuth struct{}

// Login implements the auth.Auth interface
func (g GitHubCopilotAuth) Login() (auth.OAuthCredentials, error) {
	ctx := context.Background()

	// Request device code
	dc, err := requestDeviceCode(ctx)
	if err != nil {
		return auth.OAuthCredentials{}, fmt.Errorf("failed to request device code: %w", err)
	}

	// Display user instructions
	fmt.Printf("Please visit: %s\n", dc.VerificationURI)
	fmt.Printf("And enter code: %s\n", dc.UserCode)

	// Poll for token
	token, err := pollForToken(ctx, dc)
	if err != nil {
		return auth.OAuthCredentials{}, fmt.Errorf("failed to get token: %w", err)
	}

	fmt.Printf("✓ Successfully authenticated with GitHub Copilot\n")

	return auth.OAuthCredentials{
		Access:   token.AccessToken,
		Refresh:  token.RefreshToken,
		Expires:  token.ExpiresAt,
		Metadata: map[string]string{},
	}, nil
}

// RefreshToken implements the auth.Auth interface
func (g GitHubCopilotAuth) RefreshToken(refreshToken string) (auth.OAuthCredentials, error) {
	ctx := context.Background()

	token, err := getCopilotToken(ctx, refreshToken)
	if err != nil {
		return auth.OAuthCredentials{}, fmt.Errorf("failed to refresh copilot token: %w", err)
	}

	return auth.OAuthCredentials{
		Access:   token.AccessToken,
		Refresh:  token.RefreshToken,
		Expires:  token.ExpiresAt,
		Metadata: map[string]string{},
	}, nil
}

type copilotToken struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

type DeviceCode struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// requestDeviceCode initiates the device code flow with GitHub.
func requestDeviceCode(ctx context.Context) (*DeviceCode, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("scope", "read:user")

	req, err := http.NewRequestWithContext(ctx, "POST", deviceCodeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("device code request failed: %s - %s", resp.Status, string(body))
	}

	var dc DeviceCode
	if err := json.NewDecoder(resp.Body).Decode(&dc); err != nil {
		return nil, err
	}
	return &dc, nil
}

// pollForToken polls GitHub for the access token after user authorization.
func pollForToken(ctx context.Context, dc *DeviceCode) (*copilotToken, error) {
	interval := max(dc.Interval, 5)
	deadline := time.Now().Add(time.Duration(dc.ExpiresIn) * time.Second)
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}

		token, err := tryGetToken(ctx, dc.DeviceCode)
		if err == errPending {
			continue
		}
		if err == errSlowDown {
			interval += 5
			ticker.Reset(time.Duration(interval) * time.Second)
			continue
		}
		if err != nil {
			return nil, err
		}
		return token, nil
	}

	return nil, fmt.Errorf("authorization timed out")
}

var (
	errPending  = fmt.Errorf("pending")
	errSlowDown = fmt.Errorf("slow_down")
)

func tryGetToken(ctx context.Context, deviceCode string) (*copilotToken, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("device_code", deviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	req, err := http.NewRequestWithContext(ctx, "POST", accessTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	switch result.Error {
	case "":
		if result.AccessToken == "" {
			return nil, errPending
		}
		return getCopilotToken(ctx, result.AccessToken)
	case "authorization_pending":
		return nil, errPending
	case "slow_down":
		return nil, errSlowDown
	default:
		return nil, fmt.Errorf("authorization failed: %s", result.Error)
	}
}

func getCopilotToken(ctx context.Context, githubToken string) (*copilotToken, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", copilotTokenURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", githubToken))
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Editor-Version", "vscode/1.96.0")
	req.Header.Set("Editor-Plugin-Version", "copilot-chat/0.32.4")
	req.Header.Set("Copilot-Integration-Id", "vscode-chat")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrNotAvailable
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("copilot token request failed: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Token     string `json:"token"`
		ExpiresAt int64  `json:"expires_at"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &copilotToken{
		AccessToken:  result.Token,
		RefreshToken: githubToken,
		ExpiresAt:    result.ExpiresAt,
	}, nil
}
