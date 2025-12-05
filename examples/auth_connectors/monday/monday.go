package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/examples/utils"
	"github.com/amp-labs/connectors/generic"
	"github.com/amp-labs/connectors/providers"
)

const ApiKey = "<your-api-key>" // nolint:gosec

// Run this example with `go run anthropic.go`.
func main() {
	utils.Run(mondayAuthExample)
}

// Use the auth connector to make a request to the Monday API directly.
// The constructed client will take care of certain things like authentication,
// URL construction, and response parsing. It's a thin wrapper around the
// provider's REST API.
func mondayAuthExample(ctx context.Context) error {
	// Create an auth connector
	conn := createAuthConnector(ctx)

	type Query struct {
		QueryString string `json:"query"`
	}

	// Call the Monday API
	response, err := conn.Client.Post(ctx, "/v2", &Query{
		QueryString: "query { me { is_guest created_at name id}}",
	})
	if err != nil {
		return err
	}

	// Check the response status code
	if response.Code != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.Code) // nolint:err113
	}

	body, ok := response.Body()
	if !ok {
		return fmt.Errorf("cannot get messages: %w", common.ErrEmptyJSONHTTPResponse)
	}

	// The response body is already parsed (as JSON). You can access it like this:
	nodes, err := body.JSONPath("$.model")
	if err != nil {
		return err
	}

	// Print out the model field
	fmt.Printf("model: %s\n", nodes[0].MustString()) // nolint:forbidigo

	return nil
}

// Create an auth connector with the Monday provider.
func createAuthConnector(ctx context.Context) *generic.Connector {
	conn, err := generic.NewConnector(providers.Monday,
		generic.WithAuthenticatedClient(createAuthenticatedHTTPClient(ctx)))
	if err != nil {
		panic(err)
	}

	return conn
}

// Create an api-key authenticated HTTP client for Monday.
func createAuthenticatedHTTPClient(ctx context.Context) common.AuthenticatedHTTPClient {
	info, err := providers.ReadInfo(providers.Monday, nil)
	if err != nil {
		panic(err)
	}

	return utils.CreateApiKeyClient(ctx, info, utils.ApiKeyOptions{
		ApiKey: ApiKey,
	})
}
