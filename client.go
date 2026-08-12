package vaultsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MinAPI is the oldest Vault HTTP API major this client supports.
const MinAPI = "v1"

// Client talks to a Vault base URL (no secrets logged).
type Client struct {
	BaseURL   string
	GatewayID string
	HTTP      *http.Client
	Timeout   time.Duration
}

// Configured reports whether BaseURL + GatewayID are set.
func (c *Client) Configured() bool {
	return c != nil && strings.TrimSpace(c.BaseURL) != "" && strings.TrimSpace(c.GatewayID) != ""
}

func (c *Client) httpc() *http.Client {
	if c != nil && c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}

func (c *Client) timeout() time.Duration {
	if c != nil && c.Timeout > 0 {
		return c.Timeout
	}
	return 5 * time.Second
}

func (c *Client) url(path string) string {
	return strings.TrimRight(c.BaseURL, "/") + path
}

// Health hits /healthz and returns the body + status.
func (c *Client) Health(ctx context.Context) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/healthz"), nil)
	if err != nil {
		return 0, nil, err
	}
	resp, err := c.httpc().Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, b, nil
}

// PasswordLogin runs /v1/auth/begin + /v1/auth/finish with method=password.
func (c *Client) PasswordLogin(ctx context.Context, username, password string) (ok bool, user string, err error) {
	if !c.Configured() {
		return false, "", fmt.Errorf("vault client not configured")
	}
	if username == "" || password == "" {
		return false, "", nil
	}
	if _, has := ctx.Deadline(); !has {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout())
		defer cancel()
	}
	beginBody, _ := json.Marshal(map[string]any{
		"gatewayId": c.GatewayID,
		"username":  username,
		"method":    "password",
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/v1/auth/begin"), bytes.NewReader(beginBody))
	if err != nil {
		return false, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpc().Do(req)
	if err != nil {
		return false, "", err
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode >= 500 || resp.StatusCode == 0 {
		return false, "", fmt.Errorf("vault begin status %d", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return false, "", nil
	}
	var begin struct {
		ChallengeID string `json:"challengeId"`
	}
	if err := json.Unmarshal(b, &begin); err != nil || begin.ChallengeID == "" {
		return false, "", fmt.Errorf("vault begin: bad body")
	}
	finishBody, _ := json.Marshal(map[string]any{
		"challengeId": begin.ChallengeID,
		"proof": map[string]any{
			"method":    "password",
			"password":  password,
			"gatewayId": c.GatewayID,
		},
	})
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, c.url("/v1/auth/finish"), bytes.NewReader(finishBody))
	if err != nil {
		return false, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err = c.httpc().Do(req)
	if err != nil {
		return false, "", err
	}
	b, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode >= 500 {
		return false, "", fmt.Errorf("vault finish status %d", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return false, "", nil
	}
	var out struct {
		Assertion  string `json:"assertion"`
		NeedMethod string `json:"needMethod"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return false, "", fmt.Errorf("vault finish: bad body")
	}
	if out.NeedMethod != "" || out.Assertion == "" {
		return false, "", nil
	}
	return true, strings.ToLower(strings.TrimSpace(username)), nil
}

// EnrollPackage is the redeem response (identity material for the gateway).
type EnrollPackage struct {
	GatewayID string          `json:"gatewayId,omitempty"`
	Raw       json.RawMessage `json:"-"`
}

// RedeemEnroll POSTs /v1/enroll/redeem. Does not write host files.
func (c *Client) RedeemEnroll(ctx context.Context, token string) (EnrollPackage, error) {
	var zero EnrollPackage
	if strings.TrimSpace(c.BaseURL) == "" || token == "" {
		return zero, fmt.Errorf("vaultURL and token required")
	}
	body, _ := json.Marshal(map[string]any{"token": token})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/v1/enroll/redeem"), bytes.NewReader(body))
	if err != nil {
		return zero, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpc().Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return zero, fmt.Errorf("vault enroll status %d", resp.StatusCode)
	}
	var pkg EnrollPackage
	_ = json.Unmarshal(b, &pkg)
	pkg.Raw = append(json.RawMessage(nil), b...)
	return pkg, nil
}
