package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/connector"
	"github.com/amp-labs/connectors/examples/client_utils"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/salesforce"
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

func main() {
	// Run the examples
	authConnectorExample()
	deepConnectorExample()
}

// Use the auth connector to make a request to the Salesforce API directly.
// The constructed client will take care of certain things like authentication,
// URL construction, and response parsing. It's a very thin wrapper around the
// provider's REST API.
func authConnectorExample() {
	// Create an auth connector
	conn := createAuthConnector()

	// Call the Salesforce API (limits endpoint just for example)
	response, err := conn.Client.Get(context.Background(), "/services/data/v61.0/limits/")
	if err != nil {
		panic(err)
	}

	// Check the response status code
	if response.Code != http.StatusOK {
		panic(fmt.Errorf("unexpected status code: %d", response.Code))
	}

	// The response body is already parsed (as JSON). You can access it like this:
	nodes, err := response.Body.JSONPath("$.MassEmail.Max")
	if err != nil {
		panic(err)
	}

	// Print out the mass email limit
	fmt.Printf("MassEmail.Max: %f\n", nodes[0].MustNumeric())
}

// Use the deep connector to make a request to the Salesforce API using the
// Salesforce-specific connector. This connector is more opinionated and provides
// additional functionality on top of the auth connector. It handles things
// like bulk operations, error handling, and more. It's a higher-level abstraction
// than the auth connector -- you don't need to know the Salesforce API as well
// to use it. No need to construct URLs or parse responses.
func deepConnectorExample() {
	// Create the Salesforce client
	conn := createDeepConnector()

	var allRows []common.ReadResultRow

	// Make a read request to Salesforce
	result, err := conn.Read(context.Background(), connectors.ReadParams{
		ObjectName: "Contact",
		Fields:     []string{"FirstName", "LastName", "Email"},
	})
	if err != nil {
		panic(err)
	}

	// Collect the first page of results
	allRows = result.Data

	// If there are more pages, keep reading them.
	for !result.Done {
		result, err = conn.Read(context.Background(), connectors.ReadParams{
			ObjectName: "Contact",
			Fields:     []string{"FirstName", "LastName", "Email"},
			NextPage:   result.NextPage,
		})
		if err != nil {
			panic(err)
		}

		allRows = append(allRows, result.Data...)
	}

	fmt.Printf("Result is %v\n", allRows)
}

func createAuthenticatedHttpClient() common.AuthenticatedHTTPClient {
	info, err := providers.ReadInfo(providers.Salesforce, &substitutions)
	if err != nil {
		panic(err)
	}

	return client_utils.GetOAuth2AuthorizationCodeClient(info, client_utils.OAuth2AuthCodeOptions{
		OAuth2ClientId:     OAuth2ClientId,
		OAuth2ClientSecret: OAuth2ClientSecret,
		OAuth2AccessToken:  OAuth2AccessToken,
		OAuth2RefreshToken: OAuth2RefreshToken,
		Expiry:             time.Time{},
	})
}

func createAuthConnector() *connector.Connector {
	conn, err := connector.NewConnector(providers.Salesforce,
		connector.WithCatalogSubstitutions(substitutions),
		connector.WithAuthenticatedClient(createAuthenticatedHttpClient()))
	if err != nil {
		panic(err)
	}

	return conn
}

func createDeepConnector() *salesforce.Connector {
	conn, err := salesforce.NewConnector(
		salesforce.WithAuthenticatedClient(createAuthenticatedHttpClient()),
		salesforce.WithWorkspace(Workspace))
	if err != nil {
		panic(err)
	}

	return conn
}
