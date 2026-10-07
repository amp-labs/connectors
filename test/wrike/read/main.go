package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	connTest "github.com/amp-labs/connectors/test/wrike"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetWrikeConnector(ctx)

	tests := []common.ReadParams{
		{
			ObjectName: "tasks",
			Fields:     connectors.Fields("id", "title", "status", "updatedDate"),
			PageSize:   2,
		},
		{
			// Optional fields are requested through Wrike's `fields` parameter.
			ObjectName: "tasks",
			Fields:     connectors.Fields("id", "title", "description", "attachmentCount", "responsibleIds"),
		},
		{
			ObjectName: "folders",
			Fields:     connectors.Fields("id", "title", "scope"),
		},
		{
			ObjectName: "contacts",
			Fields:     connectors.Fields("id", "firstName", "lastName", "type"),
		},
		{
			ObjectName: "customfields",
			Fields:     connectors.Fields("id", "title", "type"),
		},
		{
			ObjectName: "audit_log",
			Fields:     connectors.Fields("id", "operation", "eventDate", "objectType"),
			PageSize:   30,
		},
		{
			ObjectName: "workflows",
			Fields:     connectors.Fields("id", "name", "customStatuses"),
		},
		// Incremental: provider-side updatedDate / eventDate ranges.
		{
			ObjectName: "tasks",
			Fields:     connectors.Fields("id", "title", "updatedDate"),
			Since:      time.Now().Add(-7 * 24 * time.Hour),
		},
		{
			ObjectName: "tasks",
			Fields:     connectors.Fields("id", "title", "updatedDate"),
			Since:      time.Now().Add(time.Hour),
		},
		{
			ObjectName: "audit_log",
			Fields:     connectors.Fields("id", "operation", "eventDate"),
			Since:      time.Now().Add(-24 * time.Hour),
			PageSize:   50,
		},
	}

	for _, tt := range tests {
		slog.Info("=== Reading "+tt.ObjectName+" ===", "since", tt.Since, "fields", tt.Fields.List())
		testscenario.ReadThroughPages(ctx, conn, tt)
	}
}
