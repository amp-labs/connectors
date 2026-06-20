// Unlike other connectors, Calendar push notifications carry no record IDs, so the
// subscribe pipeline passes recordIds[0] as a verbatim RFC3339 (UTC, ms) timestamp that
// becomes the events.list updatedMin query param. This test exercises that convention:
//
//  1. Capture a checkpoint timestamp, then create an event.
//  2. GetRecordsByIds([checkpoint]) — the new event must be returned.
//  3. Delete the event.
//  4. GetRecordsByIds([checkpoint]) again — the deletion must come back with
//     status:"cancelled" (this is what showDeleted=true buys us).
//
// Run with: go run ./test/google/calendar/record/main.go
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

type dateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type createPayload struct {
	Start   dateTime `json:"start"`
	End     dateTime `json:"end"`
	Summary string   `json:"summary"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetGoogleCalendarConnector(ctx)

	// Step 1: checkpoint just before the change, in the exact format the server passes in
	// (RFC3339, UTC, milliseconds). A minute of slack absorbs minor clock skew.
	updatedMin := datautils.Time.FormatRFC3339inUTCWithMilliseconds(time.Now().Add(-time.Minute))

	summary := gofakeit.Name()
	startTime := time.Now().Add(time.Hour)
	endTime := startTime.Add(time.Hour)

	createResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "events",
		RecordData: createPayload{
			Start: dateTime{
				DateTime: datautils.Time.FormatRFC3339inUTCWithMilliseconds(startTime),
				TimeZone: "America/Los_Angeles",
			},
			End: dateTime{
				DateTime: datautils.Time.FormatRFC3339inUTCWithMilliseconds(endTime),
				TimeZone: "America/Los_Angeles",
			},
			Summary: summary,
		},
	})
	if err != nil {
		utils.Fail("error creating event", "error", err)
	}

	eventID := createResult.RecordId
	if eventID == "" {
		utils.Fail("create returned no record id")
	}

	slog.Info("Created event", "id", eventID, "summary", summary)

	// Step 2: fetch changes since the checkpoint — the new event must be present.
	rows, err := conn.GetRecordsByIds(ctx, "events", []string{updatedMin},
		[]string{"id", "summary", "status"}, nil)
	if err != nil {
		utils.Fail("error fetching records by updatedMin", "error", err)
	}

	slog.Info("Fetched changed events", "count", len(rows))
	utils.DumpJSON(rows, os.Stdout)

	if !containsEvent(rows, eventID) {
		utils.Fail("created event not returned by GetRecordsByIds", "id", eventID)
	}

	slog.Info("✓ Created event was returned by the updatedMin fetch")

	// Step 3: delete the event.
	deleteResult, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "events",
		RecordId:   eventID,
	})
	if err != nil {
		utils.Fail("error deleting event", "error", err, "id", eventID)
	}

	if !deleteResult.Success {
		utils.Fail("delete operation failed", "id", eventID)
	}

	slog.Info("Deleted event", "id", eventID)

	// Step 4: fetch again — the deletion must come back as status:"cancelled".
	rows, err = conn.GetRecordsByIds(ctx, "events", []string{updatedMin},
		[]string{"id", "summary", "status"}, nil)
	if err != nil {
		utils.Fail("error fetching records by updatedMin after delete", "error", err)
	}

	slog.Info("Fetched changed events after delete", "count", len(rows))
	utils.DumpJSON(rows, os.Stdout)

	if !isCancelled(rows, eventID) {
		utils.Fail("deleted event not returned as cancelled", "id", eventID)
	}

	slog.Info("✓ Deleted event was returned with status:\"cancelled\"")
	slog.Info("✓ Test completed successfully!")
}

func containsEvent(rows []common.ReadResultRow, id string) bool {
	for _, row := range rows {
		if row.Id == id {
			return true
		}
	}

	return false
}

func isCancelled(rows []common.ReadResultRow, id string) bool {
	for _, row := range rows {
		if row.Id != id {
			continue
		}

		status, _ := row.Fields["status"].(string)

		return status == "cancelled"
	}

	return false
}
