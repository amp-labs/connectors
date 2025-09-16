package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/breakcold"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := breakcold.GetBreakcoldConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"status", "workspaces", "members", "tags", "lists", "leads"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
