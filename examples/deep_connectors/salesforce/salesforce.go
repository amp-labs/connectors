package main

import (
	"context"
	"fmt"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/examples/utils"
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

// AccessTokenExpiry is the time when the access token expires.
// This is used to determine if the token needs to be refreshed.
// If you have an actual value, you can set it here. Otherwise
// it will be set to a day ago to force a refresh.
var AccessTokenExpiry = time.Now().Add(-24 * time.Hour)

// substitutions is a map of variables that can be used in the provider catalog.
var substitutions = map[string]string{
	"workspace": Workspace,
}

// Run this example with `go run salesforce.go`
func main() {
	utils.Run(salesforceDeepExample)
}

// Use the deep connector to make a request to the Salesforce API using the
// Salesforce-specific connector. This connector is more opinionated and provides
// additional functionality on top of the auth connector. It handles things
// like bulk operations, error handling, and more. It's a higher-level abstraction
// than the auth connector -- you don't need to know the Salesforce API as well
// to use it. No need to construct URLs or parse responses.
func salesforceDeepExample(ctx context.Context) error {
	// Create the Salesforce client
	conn := createDeepConnector(ctx)

	// Make a read request to Salesforce
	result, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "Contact",
		Fields:     []string{"FirstName", "LastName", "Email"},
	})
	if err != nil {
		return err
	}

	// Collect the first page of results
	allRows := result.Data

	// If there are more pages, keep reading them.
	for !result.Done {
		result, err = conn.Read(ctx, connectors.ReadParams{
			ObjectName: "Contact",
			Fields:     []string{"FirstName", "LastName", "Email"},
			NextPage:   result.NextPage,
		})
		if err != nil {
			return err
		}

		allRows = append(allRows, result.Data...)
	}

	fmt.Printf("Result is %v\n", allRows)

	return nil
}

// Create a deep connector with the Salesforce provider.
func createDeepConnector(ctx context.Context) *salesforce.Connector {
	conn, err := salesforce.NewConnector(
		salesforce.WithAuthenticatedClient(createAuthenticatedHttpClient(ctx)),
		salesforce.WithWorkspace(Workspace))
	if err != nil {
		panic(err)
	}

	return conn
}

// Create an OAuth2 authenticated HTTP client for Salesforce.
func createAuthenticatedHttpClient(ctx context.Context) common.AuthenticatedHTTPClient {
	info, err := providers.ReadInfo(providers.Salesforce, &substitutions)
	if err != nil {
		panic(err)
	}

	return utils.CreateOAuth2AuthorizationCodeClient(ctx, info, utils.OAuth2AuthCodeOptions{
		OAuth2ClientId:     OAuth2ClientId,
		OAuth2ClientSecret: OAuth2ClientSecret,
		OAuth2AccessToken:  OAuth2AccessToken,
		OAuth2RefreshToken: OAuth2RefreshToken,
		Expiry:             AccessTokenExpiry,
	})
}
