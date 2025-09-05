package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/avoma"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := avoma.GetAvomaConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"calls", "custom_categories", "meetings", "notes", "scorecard_evaluations", "scorecards", "smart_categories", "template", "users"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
