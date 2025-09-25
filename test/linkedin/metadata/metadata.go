package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/linkedin"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := linkedin.GetConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"adTargetingFacets", "dmpEngagementSourceTypes"})
	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	utils.DumpJSON(m, os.Stdout)
}
