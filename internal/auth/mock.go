package auth

import "fmt"

type MockAuth struct{}

func (m MockAuth) Login() (OAuthCredentials, error) {
	fmt.Printf("Logged in.\n")
	return OAuthCredentials{Access: "mock", Refresh: "mock", Expires: 99999999999999}, nil
}

func (m MockAuth) RefreshToken(refreshToken string) (OAuthCredentials, error) {
	return OAuthCredentials{Access: "mock", Refresh: "mock", Expires: 99999999999999}, nil
}
