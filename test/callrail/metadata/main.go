package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/callrail"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := callrail.NewConnector(ctx)

	_, err := conn.GetPostAuthInfo(ctx)
	if err != nil {
		utils.Fail(err.Error())
	}

	m, err := conn.ListObjectMetadata(ctx, []string{"calls", "companies", "users"})
	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	utils.DumpJSON(m, os.Stdout)
}
