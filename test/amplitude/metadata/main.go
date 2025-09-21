package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/amplitude"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := amplitude.GetAmplitudeConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"annotations"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
