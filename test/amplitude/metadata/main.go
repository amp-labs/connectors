package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/amplitude"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	ctx = common.WithAuthToken(ctx, connTest.GetAmplitudeAPIkey())

	conn := connTest.GetAmplitudeConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"annotations", "taxonomy/event", "events"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
