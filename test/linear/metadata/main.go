package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/linear"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := linear.GetLinearConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"teams", "users"})

	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
