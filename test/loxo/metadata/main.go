package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/loxo"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := loxo.GetLoxoConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"companies", "countries", "deals", "activity_types", "people", "person_events"})
	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	utils.DumpJSON(m, os.Stdout)
}
