package main

import (
	"context"
	"log"
	"os"

	connTest "github.com/amp-labs/connectors/test/phoneburner"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	conn := connTest.GetPhoneBurnerConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"tags"})
	if err != nil {
		log.Fatal("Error fetching metadata: ", err)
	}

	utils.DumpJSON(m, os.Stdout)
}
