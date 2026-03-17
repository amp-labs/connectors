package main

import (
	"context"
	"log"
	"os"

	connTest "github.com/amp-labs/connectors/test/getresponse"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	conn := connTest.GetGetResponseConnector(ctx)

	meta, err := conn.ListObjectMetadata(ctx, []string{"contacts"})
	if err != nil {
		log.Fatalf("ListObjectMetadata error: %v", err)
	}

	utils.DumpJSON(meta, os.Stdout)
}
