package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	oauthdto "urlshortener/internal/modules/oauth/dto"
)

const (
	googleTokenURL    = "https://oauth2.googleapis.com/token"
	googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
)

// GenerateState generates a random 32-byte hex string to prevent CSRF.
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ExchangeGoogleCode exchanges the OAuth authorization code for Google user information.
func ExchangeGoogleCode(ctx context.Context, httpClient *http.Client, clientID, clientSecret, redirectURL, code string) (*oauthdto.GoogleUserInfo, error) {
	v := url.Values{}
	v.Set("code", code)
	v.Set("client_id", clientID)
	v.Set("client_secret", clientSecret)
	v.Set("redirect_uri", redirectURL)
	v.Set("grant_type", "authorization_code")

	resp, err := httpClient.PostForm(googleTokenURL, v)
	if err != nil {
		return nil, fmt.Errorf("token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	userInfoReq, _ := http.NewRequestWithContext(ctx, "GET", googleUserInfoURL, nil)
	userInfoReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	userResp, err := httpClient.Do(userInfoReq)
	if err != nil {
		return nil, fmt.Errorf("user info request failed: %w", err)
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(userResp.Body)
		return nil, fmt.Errorf("user info failed with status %d: %s", userResp.StatusCode, string(body))
	}

	var info oauthdto.GoogleUserInfo
	if err := json.NewDecoder(userResp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	if !info.VerifiedEmail {
		return nil, fmt.Errorf("google email not verified")
	}

	return &info, nil
}
