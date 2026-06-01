package turnstile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Response struct {
	Success bool `json:"success"`
}

func VerifyToken(secret, token string) (bool, error) {
	data := url.Values{}
	data.Set("secret", secret)
	data.Set("response", token)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://challenges.cloudflare.com/turnstile/v0/siteverify", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return false, fmt.Errorf("failed to create turnstile request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to call turnstile: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read turnstile response: %w", err)
	}

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("failed to parse turnstile response: %w", err)
	}

	return result.Success, nil
}
