package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/ashby"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := ashby.GetAshbyConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"assessment.list", "candidate.list", "interview.list"})

	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
