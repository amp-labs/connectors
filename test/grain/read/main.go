package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	grain "github.com/amp-labs/connectors/test/grain"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := grain.GetGrainConnector(ctx)

	testscenario.ReadThroughPagesFieldsOnly(ctx, conn, common.ReadParams{
		ObjectName: "recordings",
		Fields: connectors.Fields("id", "title", "start_datetime", "source", "share_state",
			"highlights", "ai_action_items", "ai_summary", "private_notes",
			"calendar_event", "hubspot", "screenshares", "ai_template_sections"),
	})

	testscenario.ReadThroughPagesFieldsOnly(ctx, conn, common.ReadParams{
		ObjectName: "uploads",
		Fields:     connectors.Fields("id", "title", "status", "updated_datetime", "attributes"),
	})

	testscenario.ReadThroughPagesFieldsOnly(ctx, conn, common.ReadParams{
		ObjectName: "uploads",
		Fields:     connectors.Fields("id", "title", "status", "updated_datetime"),
		Since:      time.Now().Add(-7 * 24 * time.Hour),
		Until:      time.Now(),
	})

	testscenario.ReadThroughPagesFieldsOnly(ctx, conn, common.ReadParams{
		ObjectName: "hooks",
		Fields:     connectors.Fields("id", "hook_type", "hook_url", "enabled"),
	})

	testscenario.ReadThroughPagesFieldsOnly(ctx, conn, common.ReadParams{
		ObjectName: "users",
		Fields:     connectors.Fields("id"),
	})

	testscenario.ReadThroughPagesFieldsOnly(ctx, conn, common.ReadParams{
		ObjectName: "teams",
		Fields:     connectors.Fields("id", "name"),
	})

	testscenario.ReadThroughPagesFieldsOnly(ctx, conn, common.ReadParams{
		ObjectName: "meeting_types",
		Fields:     connectors.Fields("id", "name", "scope"),
	})
}
