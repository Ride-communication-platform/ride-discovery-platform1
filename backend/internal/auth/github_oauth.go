package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type GitHubOAuth struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type gitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type GitHubUserInfo struct {
	Email         string
	Name          string
	VerifiedEmail bool
}

type gitHubUserProfile struct {
	Name  string `json:"name"`
	Login string `json:"login"`
	Email string `json:"email"`
}

type gitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (g *GitHubOAuth) Enabled() bool {
	return g != nil && g.ClientID != "" && g.ClientSecret != "" && g.RedirectURI != ""
}

func (g *GitHubOAuth) AuthURL(state string) string {
	values := url.Values{}
	values.Set("client_id", g.ClientID)
	values.Set("redirect_uri", g.RedirectURI)
	values.Set("scope", "read:user user:email")
	values.Set("state", state)
	return "https://github.com/login/oauth/authorize?" + values.Encode()
}

func (g *GitHubOAuth) ExchangeCode(code string) (string, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", g.ClientID)
	form.Set("client_secret", g.ClientSecret)
	form.Set("redirect_uri", g.RedirectURI)

	request, err := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github token exchange failed with status %d", response.StatusCode)
	}

	var payload gitHubTokenResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.AccessToken == "" {
		return "", fmt.Errorf("github token exchange returned empty access token")
	}

	return payload.AccessToken, nil
}

func (g *GitHubOAuth) FetchUserInfo(accessToken string) (*GitHubUserInfo, error) {
	profileRequest, err := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	profileRequest.Header.Set("Authorization", "Bearer "+accessToken)
	profileRequest.Header.Set("Accept", "application/vnd.github+json")

	profileResponse, err := http.DefaultClient.Do(profileRequest)
	if err != nil {
		return nil, err
	}
	defer profileResponse.Body.Close()
	if profileResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github profile fetch failed with status %d", profileResponse.StatusCode)
	}

	var profile gitHubUserProfile
	if err := json.NewDecoder(profileResponse.Body).Decode(&profile); err != nil {
		return nil, err
	}

	emailRequest, err := http.NewRequest(http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return nil, err
	}
	emailRequest.Header.Set("Authorization", "Bearer "+accessToken)
	emailRequest.Header.Set("Accept", "application/vnd.github+json")

	emailResponse, err := http.DefaultClient.Do(emailRequest)
	if err != nil {
		return nil, err
	}
	defer emailResponse.Body.Close()
	if emailResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github email fetch failed with status %d", emailResponse.StatusCode)
	}

	var emails []gitHubEmail
	if err := json.NewDecoder(emailResponse.Body).Decode(&emails); err != nil {
		return nil, err
	}

	info := &GitHubUserInfo{
		Name: profile.Name,
	}
	if info.Name == "" {
		info.Name = profile.Login
	}
	if profile.Email != "" {
		info.Email = profile.Email
		info.VerifiedEmail = true
	}
	for _, candidate := range emails {
		if candidate.Primary && candidate.Verified {
			info.Email = candidate.Email
			info.VerifiedEmail = true
			break
		}
	}
	if info.Email == "" {
		for _, candidate := range emails {
			if candidate.Verified {
				info.Email = candidate.Email
				info.VerifiedEmail = true
				break
			}
		}
	}
	if info.Email == "" {
		return nil, fmt.Errorf("github account has no verified email")
	}

	return info, nil
}
