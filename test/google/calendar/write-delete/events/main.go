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

type createPayload struct {
	Start   dateTime `json:"start"`
	End     dateTime `json:"end"`
	Summary string   `json:"summary"`
}

type updatePayload struct {
	Summary string `json:"summary"`
}

type dateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
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

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"events",
		createPayload{
			Start: dateTime{
				DateTime: startTimeStr,
				TimeZone: "America/Los_Angeles",
			},
			End: dateTime{
				DateTime: endTimeStr,
				TimeZone: "America/Los_Angeles",
			},
			Summary: summaryBefore,
		},
		updatePayload{
			Summary: summaryAfter,
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "summary"),
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
