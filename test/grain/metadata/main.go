package main

import (
	"context"
	"log"
	"os"

	grain "github.com/amp-labs/connectors/test/grain"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := grain.GetGrainConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{
		"recordings",
		"uploads",
		"hooks",
		"users",
		"teams",
		"meeting_types",
	})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
