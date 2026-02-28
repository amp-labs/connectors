package common

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

// zoomInfoClient implements AuthenticatedHTTPClient with JWT-based authentication.
type ZoomInfoClient struct {
	username string // used to regenerate JWT
	password string
	jwt      string // current JWT used to make API calls
	mu       sync.RWMutex
	client   *http.Client
}

// NewZoomInfoClient creates a new zoomInfoClient with initial JWT.
func NewZoomInfoClient(ctx context.Context, client *http.Client,
	username, password string,
) (*ZoomInfoClient, error) {
	// Generate initial JWT
	jwt, err := zoominfoAuth(ctx, client, username, password)
	if err != nil {
		return nil, err
	}

	return &ZoomInfoClient{
		username: username,
		password: password,
		jwt:      jwt,
		client:   client,
	}, nil
}

// Do executes an HTTP request, adding the JWT and retrying on 401 Unauthorized.
func (c *ZoomInfoClient) Do(req *http.Request) (*http.Response, error) {
	// Clone request to avoid modifying the original
	req = req.Clone(req.Context())

	c.mu.RLock()
	req.Header.Set("Authorization", "Bearer "+c.jwt)
	c.mu.RUnlock()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	// If not unauthorized, just return immediately
	if resp.StatusCode != http.StatusUnauthorized {
		return resp, nil
	}

	resp.Body.Close() // Close the response body

	// Regenerate JWT
	newJWT, err := zoominfoAuth(req.Context(), c.client, c.username, c.password) //nolint:contextcheck
	if err != nil {
		return nil, err
	}

	// Update JWT
	c.mu.Lock()
	c.jwt = newJWT
	c.mu.Unlock()

	// Retry request with new JWT
	req = req.Clone(req.Context()) //nolint:contextcheck

	req.Header.Set("Authorization", "Bearer "+newJWT)

	return c.client.Do(req)
}

func (c *ZoomInfoClient) CloseIdleConnections() {
	c.client.CloseIdleConnections()
}

// In the ZoomInfo connector, there are two types of authentication: PKI and username/password.
// We use the second type, which is similar to basic authentication (using username and password).
// However, in addition, we must call the authenticate API to generate a JWT token.
// Authenticate API link: https://api-docs.zoominfo.com/#477888fc-8308-4645-81ca-ca7a6d7ba3d1.
func zoominfoAuth(ctx context.Context, client *http.Client, username, password string) (string, error) {
	// Request body
	reqBody := map[string]string{
		"username": username,
		"password": password,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.zoominfo.com/authenticate",
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	response, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	// Parse response JSON
	var result struct {
		JWT string `json:"jwt"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.JWT, nil
}
