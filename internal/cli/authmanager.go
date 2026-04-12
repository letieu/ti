package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/letieu/ti/internal/auth"
	"github.com/letieu/ti/internal/config"
)

type AuthManager struct {
	store       *config.Manager
	creds       map[string]auth.OAuthCredentials
	providerReg map[string]auth.Auth
}

func NewAuthManager() (*AuthManager, error) {
	store, _ := config.NewManager("", "auth.json")
	creds := make(map[string]auth.OAuthCredentials)
	providerReg := make(map[string]auth.Auth)

	err := store.Load(&creds)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		// ignore file-not-exist error
	}

	providerReg["antigravity"] = auth.AntigravityAuth{}
	providerReg["mock"] = auth.MockAuth{}

	return &AuthManager{
		store:       store,
		creds:       creds,
		providerReg: providerReg,
	}, nil
}

func (m *AuthManager) GetCreds(provider string) (auth.OAuthCredentials, error) {
	creds := m.creds[provider]

	if creds.Access == "" {
		return auth.OAuthCredentials{}, fmt.Errorf("Not have credentials")
	}

	if creds.Expired() == false {
		return creds, nil
	}

	newCreds, err := m.providerReg[provider].RefreshToken(creds.Refresh)
	if err != nil {
		return auth.OAuthCredentials{}, err
	}

	newCreds.Metadata = creds.Metadata
	m.SetCreds(provider, newCreds)
	return newCreds, nil
}

func (m *AuthManager) SetCreds(provider string, creds auth.OAuthCredentials) {
	m.creds[provider] = creds
	m.store.Save(m.creds)
}

func (m *AuthManager) Login(provider string) error {
	creds, err := m.providerReg[provider].Login()
	if err != nil {
		return err
	}

	m.SetCreds(provider, creds)
	return nil
}
