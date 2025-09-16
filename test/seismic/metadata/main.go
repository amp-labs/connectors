package main

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/test/seismic"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := seismic.GetConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"entitlementRoles", "groups", "workspaceContents", "workspaceContentVersions", "emails", "emailTemplateStaticImages"})
	if err != nil {
		return err
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)

	return nil
}
