package main

import (
	"context"
	"log"
	"os"

	salesforceMarketing "github.com/amp-labs/connectors/test/salesforcemarketing"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := salesforceMarketing.GetSalesforceMarketingConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{
		"filetransferlocations",
		"campaigns",
		"contacts/schema",
		"approvals",
	})

	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
