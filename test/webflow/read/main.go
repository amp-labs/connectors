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
	connTest "github.com/amp-labs/connectors/test/webflow"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetWebflowConnector(ctx)

	tests := []common.ReadParams{
		{
			ObjectName: "sites",
			Fields:     connectors.Fields("id", "displayName", "lastUpdated"),
		},
		{
			ObjectName: "pages",
			Fields:     connectors.Fields("id", "title", "slug", "lastUpdated"),
			PageSize:   1,
		},
		{
			ObjectName: "forms",
			Fields:     connectors.Fields("id", "displayName", "lastUpdated"),
		},
		{
			ObjectName: "assets",
			Fields:     connectors.Fields("id", "displayName", "contentType", "lastUpdated"),
			PageSize:   2,
		},
		{
			ObjectName: "collections",
			Fields:     connectors.Fields("id", "displayName", "lastUpdated"),
		},
		{
			ObjectName: "form_submissions",
			Fields:     connectors.Fields("id", "displayName", "dateSubmitted"),
		},
		{
			ObjectName: "webhooks",
			Fields:     connectors.Fields("id", "triggerType", "url"),
		},
		{
			ObjectName: "components",
			Fields:     connectors.Fields("id", "name"),
		},
		// Incremental: comments are the only list sorted newest-first server-side.
		{
			ObjectName: "comments",
			Fields:     connectors.Fields("id", "content", "lastUpdated"),
			Since:      time.Now().Add(-24 * time.Hour),
		},
	}

	for _, tt := range tests {
		slog.Info("=== Reading "+tt.ObjectName+" ===", "since", tt.Since)
		testscenario.ReadThroughPages(ctx, conn, tt)
	}
}
