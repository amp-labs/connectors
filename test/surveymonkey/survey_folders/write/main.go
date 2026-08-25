package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/surveymonkey"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetSurveyMonkeyConnector(ctx)

	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "survey_folders",
		RecordData: map[string]any{
			"title": fmt.Sprintf("Amp Integration Folder %s", gofakeit.UUID()),
		},
	})
	if err != nil {
		panic(err)
	}

	utils.DumpJSON(result, os.Stdout)
}
