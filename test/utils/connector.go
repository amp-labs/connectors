package utils

import (
	"context"
	"net/http"
	"os"

	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/salesforce"
	"github.com/amp-labs/connectors/utils"
)

func Connector(ctx context.Context) (*salesforce.Connector, error) {
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

	return salesforce.NewConnector(
		salesforce.WithClient(ctx, http.DefaultClient, cfg, tok),
		salesforce.WithWorkspace(salesforceWorkspace))
}
