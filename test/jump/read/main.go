package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/jump"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetJumpConnector(ctx)

	tests := []common.ReadParams{
		// Contacts paginate with the insertedAfter window instead of cursors.
		{
			ObjectName: "contacts",
			Fields:     connectors.Fields("id", "name", "email", "status", "insertedAt", "updatedAt"),
			PageSize:   100,
		},
		{
			ObjectName: "documents",
			Fields:     connectors.Fields("id", "name", "processingState", "updatedAt"),
		},
		{
			ObjectName: "integrations",
			Fields:     connectors.Fields("id", "service", "kind", "accountEmail", "updatedAt"),
		},
		{
			ObjectName: "meetingPreps",
			Fields:     connectors.Fields("id", "contactId", "userId", "updatedAt"),
		},
		// Cursor pagination, one record per page.
		{
			ObjectName: "meetings",
			Fields: connectors.Fields("id", "topic", "status", "source", "startedAt", "duration", "updatedAt",
				"participants"),
			PageSize: 1,
		},
		{
			ObjectName: "notes",
			Fields:     connectors.Fields("id", "subject", "meetingId", "updatedAt"),
		},
		{
			ObjectName: "pulses",
			Fields:     connectors.Fields("id", "name", "description", "updatedAt"),
		},
		{
			ObjectName: "scorecards",
			Fields:     connectors.Fields("id", "name", "format", "updatedAt", "criteria"),
		},
		{
			ObjectName: "signalDefinitions",
			Fields:     connectors.Fields("id", "key", "name", "updatedAt"),
		},
		{
			ObjectName: "tasks",
			Fields:     connectors.Fields("id", "title", "status", "due", "updatedAt", "assignee"),
		},
		// Incremental: only tasks updated between Since and Until are returned.
		{
			ObjectName: "tasks",
			Fields:     connectors.Fields("id", "title", "status", "updatedAt"),
			Since:      time.Date(2026, time.June, 23, 23, 18, 9, 145000000, time.UTC),
			Until:      time.Date(2026, time.June, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			ObjectName: "users",
			Fields:     connectors.Fields("id", "fullName", "email", "owner", "updatedAt"),
		},
	}

	for _, tt := range tests {
		slog.Info("=== Reading "+tt.ObjectName+" ===", "since", tt.Since, "until", tt.Until)
		testscenario.ReadThroughPages(ctx, conn, tt)
	}
}
