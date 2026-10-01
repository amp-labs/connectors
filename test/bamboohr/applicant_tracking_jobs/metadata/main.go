package main

import (
	"context"
	"log"
	"os"

	connTest "github.com/amp-labs/connectors/test/bamboohr"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	utils.SetupLogging()

	conn := connTest.GetBambooHRConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"applicant_tracking_jobs"})
	if err != nil {
		log.Fatal("ListObjectMetadata failed: ", err)
	}

	utils.DumpJSON(m, os.Stdout)
}
