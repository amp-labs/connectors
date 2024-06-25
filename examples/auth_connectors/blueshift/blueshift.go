package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/connector"
	"github.com/amp-labs/connectors/examples/utils"
	"github.com/amp-labs/connectors/providers"
)

const ApiKey = "<api key>"

// Run this example with `go run blueshift.go`
func main() {
	utils.Run(blueshiftAuthExample)
}

// Use the auth connector to make a request to the Blueshift API directly.
// The constructed client will take care of certain things like authentication,
// URL construction, and response parsing. It's a thin wrapper around the
// provider's REST API.
func blueshiftAuthExample(ctx context.Context) error {
	// Create an auth connector
	conn := createAuthConnector(ctx)

	// Call the Blueshift API
	response, err := conn.Client.Get(ctx, "/api/v2/campaigns.json")
	if err != nil {
		return err
	}

	// Check the response status code
	if response.Code != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.Code)
	}

	// The response body is already parsed (as JSON). You can access it like this:
	nodes, err := response.Body.JSONPath("$.campaigns[0].name")
	if err != nil {
		return err
	}

	// Print out the model field
	fmt.Printf("first campaign name: %s\n", nodes[0].MustString())

	return nil
}

// Create an auth connector with the Blueshift provider.
func createAuthConnector(ctx context.Context) *connector.Connector {
	conn, err := connector.NewConnector(providers.Blueshift,
		connector.WithAuthenticatedClient(createAuthenticatedHttpClient(ctx)))
	if err != nil {
		panic(err)
	}

	return conn
}

// Create a basic-auth authenticated HTTP client for Blueshift.
func createAuthenticatedHttpClient(ctx context.Context) common.AuthenticatedHTTPClient {
	info, err := providers.ReadInfo(providers.Blueshift, nil)
	if err != nil {
		panic(err)
	}

	// Blueshift uses basic auth, but the username is set to the API key and the password is empty.
	return utils.CreateBasicAuthClient(ctx, info, utils.BasicAuthOptions{
		User: ApiKey,
		Pass: "",
	})
}
