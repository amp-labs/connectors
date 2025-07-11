package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

const (
	italianCalendar   = "en.italian#holiday@group.v.calendar.google.com"
	colorForeground   = "#ffffff"
	colorBeforeUpdate = "#19f7f0"
	colorAfterUpdate  = "#ff69b4"
)

type payload struct {
	ID              string `json:"id"`
	ForegroundColor string `json:"foregroundColor"`
	BackgroundColor string `json:"backgroundColor"`
	Selected        bool   `json:"selected"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGoogleCalendarConnector(ctx)

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"calendarList",
		payload{
			ID:              italianCalendar,
			ForegroundColor: colorForeground,
			BackgroundColor: colorBeforeUpdate,
			Selected:        true,
		},
		payload{
			ID:              italianCalendar,
			ForegroundColor: colorForeground,
			BackgroundColor: colorAfterUpdate,
			Selected:        true,
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "backgroundColor"),
			SearchBy: testscenario.Property{
				Key:   "backgroundcolor",
				Value: colorBeforeUpdate,
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"backgroundcolor": colorAfterUpdate,
			},
		},
	)
}
