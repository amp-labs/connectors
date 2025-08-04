package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/highlevelstandard"

	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := highlevelstandard.GetHighLevelStandardConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"businesses", "calendars", "calendars/groups"})

	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	utils.DumpJSON(m, os.Stdout)
}
