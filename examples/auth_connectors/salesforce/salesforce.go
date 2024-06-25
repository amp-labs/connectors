package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/connector"
	"github.com/amp-labs/connectors/examples/example_utils"
	"github.com/amp-labs/connectors/providers"
)

const (
	// Replace these with your own values.
	Workspace          = "<workspace>"
	OAuth2ClientId     = "<client id>"
	OAuth2ClientSecret = "<client secret>"
	OAuth2AccessToken  = "<access token>"
	OAuth2RefreshToken = "<refresh token>"
)

// substitutions is a map of variables that can be used in the provider catalog.
var substitutions = map[string]string{
	"workspace": Workspace,
}

// Run this example with `go run salesforce.go`
func main() {
	example_utils.Run(salesforceAuthExample)
}

// Use the auth connector to make a request to the Salesforce API directly.
// The constructed client will take care of certain things like authentication,
// URL construction, and response parsing. It's a thin wrapper around the
// provider's REST API.
func salesforceAuthExample(ctx context.Context) error {
	// Create an auth connector
	conn := createAuthConnector()

	// Call the Salesforce API (limits endpoint just for example)
	response, err := conn.Client.Get(ctx, "/services/data/v61.0/limits/")
	if err != nil {
		return err
	}

	// Check the response status code
	if response.Code != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.Code)
	}

	// The response body is already parsed (as JSON). You can access it like this:
	nodes, err := response.Body.JSONPath("$.MassEmail.Max")
	if err != nil {
		return err
	}

	// Print out the mass email limit
	fmt.Printf("MassEmail.Max: %f\n", nodes[0].MustNumeric())

	return nil
}

// Create an auth connector with the Salesforce provider.
func createAuthConnector() *connector.Connector {
	conn, err := connector.NewConnector(providers.Salesforce,
		connector.WithCatalogSubstitutions(substitutions),
		connector.WithAuthenticatedClient(createAuthenticatedHttpClient()))
	if err != nil {
		panic(err)
	}

	return conn
}

// Create an OAuth2 authenticated HTTP client for Salesforce.
func createAuthenticatedHttpClient() common.AuthenticatedHTTPClient {
	info, err := providers.ReadInfo(providers.Salesforce, &substitutions)
	if err != nil {
		panic(err)
	}

	return example_utils.CreateOAuth2AuthorizationCodeClient(info, example_utils.OAuth2AuthCodeOptions{
		OAuth2ClientId:     OAuth2ClientId,
		OAuth2ClientSecret: OAuth2ClientSecret,
		OAuth2AccessToken:  OAuth2AccessToken,
		OAuth2RefreshToken: OAuth2RefreshToken,
		Expiry:             time.Time{},
	})
}
