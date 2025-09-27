package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/jobber"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := jobber.GetJobberConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"appAlerts", "apps"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
