package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/outplay"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := outplay.GetOutplayConnector(ctx)

	objects := []string{
		"prospect",
		"prospectaccount",
		"sequence",
		"task",
	}

	m, err := conn.ListObjectMetadata(ctx, objects)
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
