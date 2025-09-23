package common

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// NewZoominfoHTTPClient returns a new http client, with automatic Basic authentication.
// Specifically this means that the client will automatically add the Basic auth header
// to every request. The username and password are provided as arguments. Additionally
// we must call the authenitcate API to get the JWT Token.
// refer: https://api-docs.zoominfo.com/#477888fc-8308-4645-81ca-ca7a6d7ba3d1.
func NewZoominfoHTTPClient( //nolint:ireturn
	ctx context.Context,
	user, pass string,
	opts ...HeaderAuthClientOption,
) (AuthenticatedHTTPClient, error) {
	token, err := ZoominfoAuth(user, pass)
	if err != nil {
		return nil, err
	}

	return NewHeaderAuthHTTPClient(ctx, append(opts, WithHeaders(Header{
		Key:   "Authorization",
		Value: "Bearer " + token,
	}))...)
}

func ZoominfoAuth(username, password string) (string, error) {
	// Request body
	reqBody := map[string]string{
		"username": username,
		"password": password,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	client := &http.Client{}

	// Send request
	response, err := client.Post(
		"https://api.zoominfo.com/authenticate",
		"application/json",
		bytes.NewReader(jsonData),
	)
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
