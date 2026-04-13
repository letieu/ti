package cloudcodeassist

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/letieu/ti/internal/auth"
)

type CloudCodeAssistAuth struct{}

const AUTH_URL = "https://accounts.google.com/o/oauth2/v2/auth"
const REDIRECT_URI = "http://localhost:51121/oauth-callback"
const AUTH_TOKEN_URL = "https://oauth2.googleapis.com/token"

const AUTH_CLIENT_ID = "681255809395-oo8ft2oprdrnp9e3aqf6av3hmdib135j.apps.googleusercontent.com"
const AUTH_SECRET = "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"

var AUTH_SCOPES = []string{
	"https://www.googleapis.com/auth/cloud-platform",
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/userinfo.profile",
}

func (a CloudCodeAssistAuth) Login() (auth.OAuthCredentials, error) {
	verifier, challenge, err := generatePKCE()
	if err != nil {
		return auth.OAuthCredentials{}, err
	}

	result, err := a.getCode(challenge, verifier)
	if err != nil {
		return auth.OAuthCredentials{}, err
	}

	tokenRes, err := a.exchangeCodeForToken(result.Code, verifier)
	if err != nil {
		return auth.OAuthCredentials{}, err
	}

	// Get user email
	email, err := a.getUserEmail(tokenRes.AccessToken)
	if err != nil {
		return auth.OAuthCredentials{}, fmt.Errorf("failed to get user email: %w", err)
	}

	// Discover project ID
	projectId, err := discoverProject(tokenRes.AccessToken)
	if err != nil {
		return auth.OAuthCredentials{}, fmt.Errorf("failed to discover project: %w", err)
	}

	// Calculate expiry time (current time + expires_in seconds - 5 min buffer)
	expiresAt := time.Now().Unix() + int64(tokenRes.ExpiresIn) - 5*60

	return auth.OAuthCredentials{
		Access:  tokenRes.AccessToken,
		Refresh: tokenRes.RefreshToken,
		Expires: expiresAt,
		Metadata: map[string]string{
			"email":      email,
			"project_id": projectId,
		},
	}, nil
}

func (a CloudCodeAssistAuth) RefreshToken(refreshToken string) (auth.OAuthCredentials, error) {
	formData := url.Values{
		"client_id":     {AUTH_CLIENT_ID},
		"client_secret": {AUTH_SECRET},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		AUTH_TOKEN_URL,
		strings.NewReader(formData.Encode()),
	)

	if err != nil {
		return auth.OAuthCredentials{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return auth.OAuthCredentials{}, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return auth.OAuthCredentials{}, fmt.Errorf("antigravity token refresh failed: %s", string(body))
	}

	var data TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return auth.OAuthCredentials{}, err
	}

	finalRefresh := data.RefreshToken
	if finalRefresh == "" {
		finalRefresh = refreshToken
	}

	expiresAt := time.Now().Unix() + int64(data.ExpiresIn) - 5*60

	return auth.OAuthCredentials{
		Refresh:  finalRefresh,
		Access:   data.AccessToken,
		Expires:  expiresAt,
		Metadata: map[string]string{},
	}, nil
}

func (CloudCodeAssistAuth) getCode(challenge string, verifier string) (*auth.CallbackResult, error) {
	params := url.Values{
		"client_id":             {AUTH_CLIENT_ID},
		"redirect_uri":          {REDIRECT_URI},
		"access_type":           {"offline"},
		"scope":                 {strings.Join(AUTH_SCOPES, " ")},
		"code_challenge_method": {"S256"},
		"code_challenge":        {challenge},
		"state":                 {verifier},
		//
		"response_type": {"code"},
		"prompt":        {"consent"},
	}

	authURL := AUTH_URL + "?" + params.Encode()

	fmt.Printf("URL: %s", authURL)

	srv, err := auth.StartCallbackServer()
	if err != nil {
		return nil, err
	}

	defer srv.Shutdown(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result, err := srv.WaitForCode(ctx)
	return result, nil
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (CloudCodeAssistAuth) exchangeCodeForToken(code string, verifier string) (TokenResponse, error) {
	params := url.Values{
		"client_id":     {AUTH_CLIENT_ID},
		"client_secret": {AUTH_SECRET},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {REDIRECT_URI},
		"code_verifier": {verifier},
	}

	req, err := http.NewRequest("POST", AUTH_TOKEN_URL, bytes.NewBufferString(params.Encode()))
	if err != nil {
		return TokenResponse{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return TokenResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return TokenResponse{}, fmt.Errorf("Error resp: %s", resp.Status)
	}

	var token TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&token)

	if err != nil {
		return TokenResponse{}, err
	}

	return token, nil
}

func generatePKCE() (verifier string, challenge string, err error) {
	// Generate random verifier
	verifierBytes := make([]byte, 32)
	if _, err = rand.Read(verifierBytes); err != nil {
		return
	}
	verifier = base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Compute SHA-256 challenge
	hash := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(hash[:])
	return
}

type UserInfoResponse struct {
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
}

// getUserEmail fetches the user's email from Google's userinfo endpoint
func (CloudCodeAssistAuth) getUserEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get user info: %s - %s", resp.Status, string(body))
	}

	var userInfo UserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", err
	}

	return userInfo.Email, nil
}

const DEFAULT_PROJECT_ID = "rising-fact-p41fc"

type LoadCodeAssistResponse struct {
	CloudAICompanionProject any `json:"cloudaicompanionProject"`
}

func discoverProject(
	accessToken string,
) (string, error) {

	endpoint := "https://cloudcode-pa.googleapis.com"

	bodyPayload := map[string]any{
		"metadata": map[string]string{
			"ideType":    "IDE_UNSPECIFIED",
			"platform":   "PLATFORM_UNSPECIFIED",
			"pluginType": "GEMINI",
		},
	}

	bodyBytes, err := json.Marshal(bodyPayload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		endpoint+"/v1internal:loadCodeAssist",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "google-api-nodejs-client/9.15.1")
	req.Header.Set("X-Goog-Api-Client", "google-cloud-sdk vscode_cloudshelleditor/0.1")
	req.Header.Set("Client-Metadata", `{"ideType":"IDE_UNSPECIFIED","platform":"PLATFORM_UNSPECIFIED","pluginType":"GEMINI"}`)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var parsed LoadCodeAssistResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}

	// Case 1: string
	if str, ok := parsed.CloudAICompanionProject.(string); ok && str != "" {
		return str, nil
	}

	// Case 2: object { id }
	if obj, ok := parsed.CloudAICompanionProject.(map[string]any); ok {
		if id, ok := obj["id"].(string); ok && id != "" {
			return id, nil
		}
	}

	// fallback
	return DEFAULT_PROJECT_ID, nil
}
