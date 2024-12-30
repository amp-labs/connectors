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

const (
	// Replace these with your own values.
	OAuth2ClientId     = "<client id>"
	OAuth2ClientSecret = "<client secret>"
)

// Run this example with `go run adobe.go`.
func main() {
	utils.Run(adobeAuthExample)
}

// Use the auth connector to make a request to the Adobe API directly.
// The constructed client will take care of certain things like authentication,
// URL construction, and response parsing. It's a thin wrapper around the
// provider's REST API.
func adobeAuthExample(ctx context.Context) error {
	// Create an auth connector
	conn := createAuthConnector(ctx)

	// Call the Adobe API (limits endpoint just for example)
	response, err := conn.Client.Get(ctx, "/data/foundation/catalog/dataSets")
	if err != nil {
		return err
	}

	// Check the response status code
	if response.Code != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.Code)
	}

	body, ok := response.Body()
	if !ok {
		return fmt.Errorf("empty response: %w", common.ErrEmptyJSONHTTPResponse)
	}

	// The response body is already parsed (as JSON). You can access it like this:
	nodes, err := body.JSONPath("$.*.name")
	if err != nil {
		return err
	}

	// Print out the names of the data sets
	for _, node := range nodes {
		fmt.Printf("DataSet.Name: %s\n", node.MustString())
	}

	return nil
}

// Create an auth connector with the Adobe provider.
func createAuthConnector(ctx context.Context) *connector.Connector {
	conn, err := connector.NewConnector(providers.Adobe,
		connector.WithAuthenticatedClient(createAuthenticatedHttpClient(ctx)))
	if err != nil {
		panic(err)
	}

	return conn
}

// Create an OAuth2 authenticated HTTP client for Adobe.
func createAuthenticatedHttpClient(ctx context.Context) common.AuthenticatedHTTPClient {
	info, err := providers.ReadInfo(providers.Adobe, nil)
	if err != nil {
		panic(err)
	}

	return utils.CreateOAuth2ClientCredentialsClient(ctx, info, utils.OAuth2ClientCredentialsOptions{
		OAuth2ClientId:     OAuth2ClientId,
		OAuth2ClientSecret: OAuth2ClientSecret,
	})
}
