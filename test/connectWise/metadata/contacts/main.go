package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/connectWise"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := connectWise.GetConnectWiseConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{"contacts"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(metadata, os.Stdout)
}
