package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/calendly"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

// Values of the events that exist in the test account. There are three of them, a day apart:
//   - bf1b889a-..., starts 2026-11-30T06:00:00Z, invitee search.alpha@example.com, active.
//   - d6350260-..., starts 2026-12-01T06:00:00Z, invitee search.beta@example.com, active.
//   - 4b185cc0-..., starts 2026-12-02T06:00:00Z, invitee search.alpha@example.com, canceled.
const (
	eventUUID    = "bf1b889a-fecc-492d-a5d5-8d224bf5b1f9"
	inviteeEmail = "search.alpha@example.com"
	eventStatus  = "canceled"
	minStartTime = "2026-11-30T00:00:00Z"
)

// nolint:gochecknoglobals
var maxStartTime = time.Date(2026, 12, 1, 6, 0, 0, 0, time.UTC) // start of the second event, the bound is inclusive.

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetCalendlyConnector(ctx)

	// Organization and user of the connection scope the list of events.
	if _, err := conn.GetPostAuthInfo(ctx); err != nil {
		utils.Fail("error retrieving post auth info", "error", err)
	}

	fields := connectors.Fields("uri", "name", "status", "start_time")

	slog.Info("Searching event by uuid...")
	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "scheduled_events",
		Fields:     fields,
		Filter:     equals("uuid", eventUUID),
	})

	slog.Info("Searching events by invitee email...")
	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "scheduled_events",
		Fields:     fields,
		Filter:     equals("invitee_email", inviteeEmail),
		Limit:      1,
	})

	slog.Info("Searching events by status...")
	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "scheduled_events",
		Fields:     fields,
		Filter:     equals("status", eventStatus),
	})

	slog.Info("Searching events by start time range...")
	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "scheduled_events",
		Fields:     fields,
		Filter:     equals("min_start_time", minStartTime, "max_start_time", maxStartTime),
	})

	slog.Info("Searching events by invitee email, status and start time...")
	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "scheduled_events",
		Fields:     fields,
		Filter:     equals("invitee_email", inviteeEmail, "status", "active", "min_start_time", minStartTime),
	})
}

// equals builds a filter out of field name and value pairs.
func equals(pairs ...any) connectors.SearchFilter {
	filter := connectors.SearchFilter{}

	for index := 0; index+1 < len(pairs); index += 2 {
		fieldName, _ := pairs[index].(string)

		filter.FieldFilters = append(filter.FieldFilters, common.FieldFilter{
			FieldName: fieldName,
			Operator:  common.FilterOperatorEQ,
			Value:     pairs[index+1],
		})
	}

	return filter
}
