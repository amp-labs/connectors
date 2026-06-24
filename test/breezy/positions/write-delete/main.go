package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/breezy"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetBreezyConnector(ctx)

	title := fmt.Sprintf("Amp integration %s", gofakeit.Word())

	createPayload := map[string]any{
		"name":        title,
		"type":        "fullTime",
		"description": "Temporary connector integration test position.",
		"location": map[string]any{
			"country":   "US",
			"state":     "CA",
			"city":      "San Francisco",
			"is_remote": true,
		},
		"department":  "Engineering",
		"category":    "software",
		"experience":  "mid-level",
		"pipeline_id": "default",
	}

	updatePayload := map[string]any{
		"name":        title + " (Updated)",
		"type":        "fullTime",
		"description": "Updated via connector integration test.",
		"location": map[string]any{
			"country":   "US",
			"state":     "CA",
			"city":      "San Francisco",
			"is_remote": true,
		},
		"department":  "Engineering",
		"category":    "software",
		"experience":  "senior-level",
		"pipeline_id": "default",
	}

	testscenario.ValidateCreateUpdateDelete(ctx, conn, "positions", createPayload, updatePayload,
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("_id", "name", "state"),
			RecordIdentifierKey: "_id",
			UpdatedFields: map[string]string{
				"name": title + " (Updated)",
			},
		},
	)
}
