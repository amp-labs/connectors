package common

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

// zoomInfoClient implements AuthenticatedHTTPClient with JWT-based authentication
type ZoomInfoClient struct {
	username string // used to regenerate JWT
	password string
	jwt      string // current JWT used to make API calls
	mu       sync.RWMutex
	client   *http.Client
}

// NewZoomInfoClient creates a new zoomInfoClient with initial JWT
func NewZoomInfoClient(ctx context.Context, client *http.Client,
	username, password string,
) (*ZoomInfoClient, error) {
	// Generate initial JWT
	jwt, err := ZoominfoAuth(ctx, client, username, password)
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

// Do executes an HTTP request, adding the JWT and retrying on 401 Unauthorized
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

	// Check for 401 Unauthorized
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close() // Close the response body

		// Regenerate JWT
		newJWT, err := ZoominfoAuth(req.Context(), c.client, c.username, c.password)
		if err != nil {
			return nil, err
		}

		// Update JWT
		c.mu.Lock()
		c.jwt = newJWT
		c.mu.Unlock()

		// Retry request with new JWT
		req = req.Clone(req.Context())

		req.Header.Set("Authorization", "Bearer "+newJWT)

		resp, err = c.client.Do(req)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func (c *ZoomInfoClient) CloseIdleConnections() {
	c.client.CloseIdleConnections()
}

func ZoominfoAuth(ctx context.Context, client *http.Client, username, password string) (string, error) {
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
