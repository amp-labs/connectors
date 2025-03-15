package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/github"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := github.GetGithubConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"deliveries", "installation/repositories"})

	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
