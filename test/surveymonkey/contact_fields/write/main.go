package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/surveymonkey"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetSurveyMonkeyConnector(ctx)

	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "contact_fields",
		Fields:     connectors.Fields("id", "label"),
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(readResult.Data) == 0 {
		log.Fatal("no contact fields found to update")
	}

	fieldID := readResult.Data[0].Id

	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contact_fields",
		RecordId:   fieldID,
		RecordData: map[string]any{
			"label": fmt.Sprintf("Amp Integration %s", fieldID),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(result, os.Stdout)
}
