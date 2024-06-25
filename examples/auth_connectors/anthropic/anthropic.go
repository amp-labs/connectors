package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/connector"
	"github.com/amp-labs/connectors/examples/example_utils"
	"github.com/amp-labs/connectors/providers"
)

const ApiKey = "<api key>"

// Run this example with `go run anthropic.go`
func main() {
	example_utils.Run(anthropicAuthExample)
}

// Use the auth connector to make a request to the Anthropic API directly.
// The constructed client will take care of certain things like authentication,
// URL construction, and response parsing. It's a thin wrapper around the
// provider's REST API.
func anthropicAuthExample(ctx context.Context) error {
	// Create an auth connector
	conn := createAuthConnector(ctx)

	type Message struct {
		Model     string            `json:"model"`
		MaxTokens int               `json:"max_tokens"`
		Messages  map[string]string `json:"messages"`
	}

	// Call the Anthropic API
	response, err := conn.Client.Post(ctx, "/v1/messages", &Message{
		Model:     "claude-3-5-sonnet-20240620",
		MaxTokens: 1024,
		Messages: map[string]string{
			"role":    "user",
			"content": "Hello, Claude",
		},
	})
	if err != nil {
		return err
	}

	// Check the response status code
	if response.Code != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.Code)
	}

	// The response body is already parsed (as JSON). You can access it like this:
	nodes, err := response.Body.JSONPath("$.model")
	if err != nil {
		return err
	}

	// Print out the model field
	fmt.Printf("model: %s\n", nodes[0].MustString())

	return nil
}

// Create an auth connector with the Anthropic provider.
func createAuthConnector(ctx context.Context) *connector.Connector {
	conn, err := connector.NewConnector(providers.Anthropic,
		connector.WithAuthenticatedClient(createAuthenticatedHttpClient(ctx)))
	if err != nil {
		panic(err)
	}

	return conn
}

// Create an api-key authenticated HTTP client for Anthropic.
func createAuthenticatedHttpClient(ctx context.Context) common.AuthenticatedHTTPClient {
	info, err := providers.ReadInfo(providers.Anthropic, nil)
	if err != nil {
		panic(err)
	}

	return example_utils.CreateApiKeyClient(ctx, info, example_utils.ApiKeyOptions{
		ApiKey: ApiKey,
	})
}
