package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type updatePayload struct {
	Summary string `json:"summary"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGoogleCalendarConnector(ctx)

	summaryBefore := gofakeit.Name()
	summaryAfter := gofakeit.Name()

	startTime := time.Now().Add(time.Hour)
	endTime := startTime.Add(time.Hour)
	startTimeStr := datautils.Time.FormatRFC3339inUTCWithMilliseconds(startTime)
	endTimeStr := datautils.Time.FormatRFC3339inUTCWithMilliseconds(endTime)

	// Timing, origin and tracing are sent as the dotted field names exposed by the
	// metadata schema; the connector nests them into the objects Google expects.
	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"events",
		map[string]any{
			"summary":                          summaryBefore,
			"start.dateTime":                   startTimeStr,
			"start.timeZone":                   "America/Los_Angeles",
			"end.dateTime":                     endTimeStr,
			"end.timeZone":                     "America/Los_Angeles",
			"source.title":                     "Ampersand live test",
			"source.url":                       "https://withampersand.com",
			"extendedProperties.private.crmId": "42",
		},
		updatePayload{
			Summary: summaryAfter,
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "summary", "start.dateTime", "source.title",
				"extendedProperties.private.crmId"),
			SearchBy: testscenario.Property{
				Key:   "summary",
				Value: summaryBefore,
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"summary": summaryAfter,
			},
		},
	)
}
