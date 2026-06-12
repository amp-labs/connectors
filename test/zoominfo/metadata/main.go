package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zoominfo"
)

func main() {
	ctx := context.Background()

	conn := connTest.GetZoomInfoConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"products", "contacts", "companies", "news", "scoops"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
