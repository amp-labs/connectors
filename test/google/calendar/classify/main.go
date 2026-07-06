// Live test for Calendar subscription-event classification — the step the subscribe pipeline
// runs on the rows GetRecordsByIds returns. It exercises all three event types against a real
// calendar and checks that CalendarSubscriptionEventsFromRecords labels each one correctly.
//
// To get a create, an update and a delete on the same fetch, we straddle a checkpoint:
//  1. Create the update and delete targets before the checkpoint.
//  2. Capture the checkpoint (updatedMin). The sleeps on either side keep the "before" events
//     and the "after" changes clearly apart even if the local and Google clocks drift a little.
//  3. After the checkpoint, create a third event (the create), edit the update target (an
//     update: created before, changed after) and delete the delete target.
//  4. Fetch from the checkpoint, classify the rows, and check each event came out as expected.
//
// Run with: go run ./test/google/calendar/classify/main.go
package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/google"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

// separation keeps the pre-checkpoint events and post-checkpoint changes on opposite sides of
// the checkpoint even with some clock skew between here and Google.
const separation = 5 * time.Second

type dateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type eventPayload struct {
	Start   dateTime `json:"start"`
	End     dateTime `json:"end"`
	Summary string   `json:"summary"`
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetGoogleCalendarConnector(ctx)

	// Step 1: the "old" events, created before the checkpoint.
	updateID := createEvent(ctx, conn, "classify-update-"+gofakeit.UUID())
	deleteID := createEvent(ctx, conn, "classify-delete-"+gofakeit.UUID())
	slog.Info("Created pre-checkpoint events", "updateTarget", updateID, "deleteTarget", deleteID)

	// Step 2: checkpoint, fenced by sleeps so the old events land before it.
	time.Sleep(separation)
	checkpoint := datautils.Time.FormatRFC3339inUTCWithMilliseconds(time.Now())
	slog.Info("Captured checkpoint (updatedMin)", "checkpoint", checkpoint)
	time.Sleep(separation)

	// Steps 3-5: the changes, all after the checkpoint.
	createID := createEvent(ctx, conn, "classify-create-"+gofakeit.UUID())
	slog.Info("Created post-checkpoint event", "createTarget", createID)

	updateEvent(ctx, conn, updateID, "classify-update-edited-"+gofakeit.UUID())
	slog.Info("Updated event", "id", updateID)

	deleteEvent(ctx, conn, deleteID)
	slog.Info("Deleted event", "id", deleteID)

	// Step 6: fetch the window and classify.
	rows, err := conn.GetRecordsByIds(ctx, "events", []string{checkpoint},
		[]string{"id", "summary", "status", "created", "updated"}, nil)
	if err != nil {
		utils.Fail("error fetching records by updatedMin", "error", err)
	}

	slog.Info("Fetched changed events", "count", len(rows))

	events := google.CalendarSubscriptionEventsFromRecords(rows, checkpoint)

	typeByID := make(map[string]common.SubscriptionEventType, len(events))

	for _, evt := range events {
		id, err := evt.RecordId()
		if err != nil {
			utils.Fail("error reading record id", "error", err)
		}

		eventType, err := evt.EventType()
		if err != nil {
			utils.Fail("error classifying event", "error", err, "id", id)
		}

		typeByID[id] = eventType
		slog.Info("Classified", "id", id, "type", eventType)
	}

	// Cleanup the two events left behind (the delete target is already gone).
	defer func() {
		deleteEvent(ctx, conn, createID)
		deleteEvent(ctx, conn, updateID)
		slog.Info("Cleaned up create/update targets")
	}()

	expect(typeByID, createID, common.SubscriptionEventTypeCreate)
	expect(typeByID, updateID, common.SubscriptionEventTypeUpdate)
	expect(typeByID, deleteID, common.SubscriptionEventTypeDelete)

	slog.Info("✓ Test completed successfully!")
}

// expect fails the run unless the classified type for id matches want.
func expect(typeByID map[string]common.SubscriptionEventType, id string, want common.SubscriptionEventType) {
	got, ok := typeByID[id]
	if !ok {
		utils.Fail("event not returned by the updatedMin fetch", "id", id, "want", want)
	}

	if got != want {
		utils.Fail("event classified incorrectly", "id", id, "want", want, "got", got)
	}

	slog.Info("✓ Correctly classified", "id", id, "type", want)
}

func createEvent(ctx context.Context, conn *google.Connector, summary string) string {
	start := time.Now().Add(time.Hour)
	end := start.Add(time.Hour)

	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "events",
		RecordData: eventPayload{
			Start:   dateTime{DateTime: datautils.Time.FormatRFC3339inUTCWithMilliseconds(start), TimeZone: "America/Los_Angeles"},
			End:     dateTime{DateTime: datautils.Time.FormatRFC3339inUTCWithMilliseconds(end), TimeZone: "America/Los_Angeles"},
			Summary: summary,
		},
	})
	if err != nil {
		utils.Fail("error creating event", "error", err, "summary", summary)
	}

	if result.RecordId == "" {
		utils.Fail("create returned no record id", "summary", summary)
	}

	return result.RecordId
}

// updateEvent patches an existing event. Calendar update replaces the resource, so start/end
// are resent alongside the new summary to keep the event valid.
func updateEvent(ctx context.Context, conn *google.Connector, id, summary string) {
	start := time.Now().Add(time.Hour)
	end := start.Add(time.Hour)

	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "events",
		RecordId:   id,
		RecordData: eventPayload{
			Start:   dateTime{DateTime: datautils.Time.FormatRFC3339inUTCWithMilliseconds(start), TimeZone: "America/Los_Angeles"},
			End:     dateTime{DateTime: datautils.Time.FormatRFC3339inUTCWithMilliseconds(end), TimeZone: "America/Los_Angeles"},
			Summary: summary,
		},
	})
	if err != nil {
		utils.Fail("error updating event", "error", err, "id", id)
	}

	if !result.Success {
		utils.Fail("update operation failed", "id", id)
	}
}

func deleteEvent(ctx context.Context, conn *google.Connector, id string) {
	result, err := conn.Delete(ctx, common.DeleteParams{ObjectName: "events", RecordId: id})
	if err != nil {
		utils.Fail("error deleting event", "error", err, "id", id)
	}

	if !result.Success {
		utils.Fail("delete operation failed", "id", id)
	}
}
