package main

import (
	"context"
	"log"
	"os"

	connTest "github.com/amp-labs/connectors/test/surveymonkey"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := connTest.GetSurveyMonkeyConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"contact_lists"})
	if err != nil {
		log.Fatal("Error fetching metadata: ", err)
	}

	utils.DumpJSON(m, os.Stdout)
}
