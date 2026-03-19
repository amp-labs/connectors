package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	devrevtest "github.com/amp-labs/connectors/test/devrev"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := devrevtest.GetConnector(ctx)
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "accounts",
		Fields:     connectors.Fields("display_id"),
		PageSize:   3,
		Since:      time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC),
	})
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "articles",
		Fields:     connectors.Fields("display_id"),
		PageSize:   3,
		Since:      time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
	})
	slog.Info("read objects test")
	readObjectsTest(ctx, conn)
}

// readObjectsTest smoke-tests every list object (1 record for those that support limit) to verify endpoints.
func readObjectsTest(ctx context.Context, conn connectors.ReadConnector) {
	since := time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC)

	for _, objectName := range readObjects {
		_, err := conn.Read(ctx, connectors.ReadParams{
			ObjectName: objectName,
			Fields:     connectors.Fields("id"),
			PageSize:   1,
			Since:      since,
		})
		if err != nil {
			slog.Error("read failed", "object", objectName, "error", err)
			continue
		}
		slog.Info("success", "object", objectName)
	}

}

// DevRev objects that will be supported by the connector for reading.
var readObjects = []string{
	"accounts",
	"articles",
	"auth-tokens",
	"code-changes",
	"commands",
	"content-template",
	"conversations",
	"dev-users",
	"directories",
	"engagements",
	"groups",
	"incidents",
	"jobs",
	"link-types.custom",
	"meetings",
	"metric-definitions",
	"org-schedules",
	"parts",
	"question-answers",
	"rev-orgs",
	"rev-users",
	"schemas.custom",
	"schemas.stock",
	"schemas.subtypes",
	"sla-trackers",
	"slas",
	"stage-diagrams",
	"stages.custom",
	"states.custom",
	"surveys",
	"surveys.responses",
	"sys-users",
	"tags",
	"vistas",
	"vistas.groups",
	"webhooks",
	"works",
}
