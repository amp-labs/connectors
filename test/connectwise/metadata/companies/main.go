package main

import (
	"context"
	"log"
	"os"

	connTest "github.com/amp-labs/connectors/test/connectwise"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := connTest.GetConnectWiseConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{"companies"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(metadata, os.Stdout)
}
