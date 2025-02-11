package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/google"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

var objectName = "calendarList"

const (
	italianCalendar   = "en.italian#holiday@group.v.calendar.google.com"
	colorBeforeUpdate = "#19f7f0"
	colorAfterUpdate  = "#ff69b4"
	colorForeground   = "#ffffff"
)

type calendarListPayload struct {
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

	conn := connTest.GetGoogleConnector(ctx, google.ModuleCalendar)

	slog.Info("> TEST Create/Update/Delete calendarList")
	slog.Info("Creating calendarList")

	createCalendarList(ctx, conn, &calendarListPayload{
		ID:              italianCalendar,
		ForegroundColor: colorForeground,
		BackgroundColor: colorBeforeUpdate,
		Selected:        true,
	})

	slog.Info("Reading calendarList")

	res := readCalendarLists(ctx, conn)

	slog.Info("Finding recently created calendarList")

	calendarList := searchCalendarList(res, "backgroundcolor", colorBeforeUpdate)
	calendarListID := fmt.Sprintf("%v", calendarList["id"])

	slog.Info("Updating calendarList name")

	updateCalendarList(ctx, conn, calendarListID, &calendarListPayload{
		ID:              italianCalendar,
		ForegroundColor: colorForeground,
		BackgroundColor: colorAfterUpdate,
		Selected:        true,
	})

	slog.Info("View that calendarList has changed accordingly")

	res = readCalendarLists(ctx, conn)

	calendarList = searchCalendarList(res, "id", calendarListID)

	calendarListSummary, ok := calendarList["backgroundcolor"].(string)
	if !ok || calendarListSummary != colorAfterUpdate {
		utils.Fail("error updated backgroundColor doesn't match")
	}

	slog.Info("Removing this calendarList")
	removeCalendarList(ctx, conn, calendarListID)
	slog.Info("> Successful test completion")
}

func searchCalendarList(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if mockutils.DoesObjectCorrespondToString(data.Fields[key], value) {
			return data.Fields
		}
	}

	utils.Fail("error finding calendarList")

	return nil
}

func readCalendarLists(ctx context.Context, conn *google.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields: connectors.Fields(
			"id", "backgroundColor",
		),
	})
	if err != nil {
		utils.Fail("error reading from google", "error", err)
	}

	return res
}

func createCalendarList(ctx context.Context, conn *google.Connector, payload *calendarListPayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to google", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a calendarList")
	}
}

func updateCalendarList(ctx context.Context, conn *google.Connector, calendarListID string, payload *calendarListPayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   calendarListID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to google", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update a calendarList")
	}
}

func removeCalendarList(ctx context.Context, conn *google.Connector, calendarListID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   calendarListID,
	})
	if err != nil {
		utils.Fail("error deleting for google", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a calendarList")
	}
}
