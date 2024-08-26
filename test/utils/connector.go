package utils

import (
	"context"
	"net/http"
	"os"

	"github.com/amp-labs/connectors/common/scanning"
	salesforce2 "github.com/amp-labs/connectors/providers/salesforce"
	"github.com/amp-labs/connectors/utils"
)

func Connector(ctx context.Context) (*salesforce2.Connector, error) {
	// assumes that this code is being run from the root of the project
	// go run test/salesforce/bulkwrite/main.go
	filePath := os.Getenv("SALESFORCE_CRED_FILE_PATH")
	if filePath == "" {
		filePath = "./salesforce-creds.json"
	}

	ampConnectionSchemaReader := JSONFileReaders(filePath)
	credentialsRegistry := scanning.NewRegistry()
	credentialsRegistry.AddReaders(ampConnectionSchemaReader...)
	salesforceWorkspace := credentialsRegistry.MustString(utils.WorkspaceRef)

	cfg := utils.SalesforceOAuthConfigFromRegistry(credentialsRegistry)
	tok := utils.SalesforceOauthTokenFromRegistry(credentialsRegistry)

	return salesforce2.NewConnector(
		salesforce2.WithClient(ctx, http.DefaultClient, cfg, tok),
		salesforce2.WithWorkspace(salesforceWorkspace))
}
