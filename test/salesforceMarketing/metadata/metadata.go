package main

import (
	"context"
	"fmt"
	"log"
	"os"

	salesforceMarketing "github.com/amp-labs/connectors/test/salesforcemarketing"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := salesforceMarketing.GetSalesforceMarketingConnector(ctx)

	// Test listing metadata for various objects from the OpenAPI spec
	m, err := conn.ListObjectMetadata(ctx, []string{
		"assets",
		"campaigns",
		"emailDefinitions",
		"smsDefinitions",
	})

	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	fmt.Println("Results:")
	utils.DumpJSON(m, os.Stdout)
}
