package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/surveymonkey"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetSurveyMonkeyConnector(ctx)

	var expectedLabel string

	updatePayload := map[string]any{}

	testscenario.ValidateUpdate(ctx, conn, "contact_fields", updatePayload,
		testscenario.UpdateTestSuite{
			ReadFields:            datautils.NewSet("id", "label"),
			RecordIdentifierKey:   "id",
			RestoreOriginalFields: true,
			PreprocessUpdatePayload: func(record *common.ReadResultRow) any {
				expectedLabel = fmt.Sprintf("Amp Integration %s", record.Id)

				return map[string]any{"label": expectedLabel}
			},
			ValidateUpdatedFields: func(record map[string]any) {
				if !mockutils.DoesObjectCorrespondToString(record["label"], expectedLabel) {
					utils.Fail("error updated properties do not match", "label", expectedLabel, record["label"])
				}
			},
		},
	)
}
