package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/fireflies"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := fireflies.GetFirefliesConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"users", "transcripts", "bites", "userGroups", "activeMeetings", "analytics"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
