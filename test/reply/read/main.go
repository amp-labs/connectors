package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/reply"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := reply.GetReplyConnector(ctx)

	// Plain read across every supported object.
	for _, objectName := range []string{
		"contacts", "contact-accounts", "contact-lists", "contact-account-lists",
		"sequences", "sequence-folders", "tasks",
		"email-templates", "email-template-folders", "email-accounts",
		"custom-fields", "schedules", "linkedin-accounts", "inbox/threads",
	} {
		readObject(ctx, conn, common.ReadParams{
			ObjectName: objectName,
			Fields:     connectors.Fields("id"),
		})
	}

	// Paginated envelope object with custom field resolution.
	readObject(ctx, conn, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("email", "firstName", "Deal Stage"),
	})

	// Root-level array object: single page.
	readObject(ctx, conn, common.ReadParams{
		ObjectName: "schedules",
		Fields:     connectors.Fields("name", "timezoneId"),
	})

	// Incremental read: only sequences support a provider-side time filter.
	readObject(ctx, conn, common.ReadParams{
		ObjectName: "sequences",
		Fields:     connectors.Fields("name", "created"),
		Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})

	// Since in the future must yield zero records, proving the filter applies.
	readObject(ctx, conn, common.ReadParams{
		ObjectName: "sequences",
		Fields:     connectors.Fields("name"),
		Since:      time.Now().Add(24 * time.Hour),
	})

	// Until has no provider-side parameter; the window is enforced
	// connector-side, so a narrow window returns only the older sequence.
	readObject(ctx, conn, common.ReadParams{
		ObjectName: "sequences",
		Fields:     connectors.Fields("name", "created"),
		Since:      time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC),
		Until:      time.Date(2026, 9, 8, 13, 0, 0, 0, time.UTC),
	})
}

func readObject(ctx context.Context, conn connectors.ReadConnector, params common.ReadParams) {
	res, err := conn.Read(ctx, params)
	if err != nil {
		utils.Fail("error reading from Reply", "object", params.ObjectName, "error", err)
	}

	slog.Info("Read result", "object", params.ObjectName,
		"since", params.Since, "rows", res.Rows, "done", res.Done)
	utils.DumpJSON(res, os.Stdout)
}
