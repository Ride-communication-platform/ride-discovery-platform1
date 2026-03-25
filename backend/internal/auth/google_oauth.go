package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type GoogleOAuth struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type GoogleUserInfo struct {
	Email         string `json:"email"`
	Name          string `json:"name"`
	VerifiedEmail bool   `json:"verified_email"`
}

func (g *GoogleOAuth) Enabled() bool {
	return g != nil && g.ClientID != "" && g.ClientSecret != "" && g.RedirectURI != ""
}

func (g *GoogleOAuth) AuthURL(state string) string {
	values := url.Values{}
	values.Set("client_id", g.ClientID)
	values.Set("redirect_uri", g.RedirectURI)
	values.Set("response_type", "code")
	values.Set("scope", "openid email profile")
	values.Set("access_type", "offline")
	values.Set("prompt", "consent")
	values.Set("state", state)
	return "https://accounts.google.com/o/oauth2/v2/auth?" + values.Encode()
}

func (g *GoogleOAuth) ExchangeCode(code string) (string, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", g.ClientID)
	form.Set("client_secret", g.ClientSecret)
	form.Set("redirect_uri", g.RedirectURI)
	form.Set("grant_type", "authorization_code")

	response, err := http.Post(
		"https://oauth2.googleapis.com/token",
		"application/x-www-form-urlencoded",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google token exchange failed with status %d", response.StatusCode)
	}

	var payload googleTokenResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.AccessToken == "" {
		return "", fmt.Errorf("google token exchange returned empty access token")
	}

	return payload.AccessToken, nil
}

func (g *GoogleOAuth) FetchUserInfo(accessToken string) (*GoogleUserInfo, error) {
	request, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google user info failed with status %d", response.StatusCode)
	}

	var payload GoogleUserInfo
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Email == "" {
		return nil, fmt.Errorf("google user info missing email")
	}

	return &payload, nil
}

func GenerateStateToken() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
