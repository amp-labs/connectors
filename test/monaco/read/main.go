package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/monaco"
	connTest "github.com/amp-labs/connectors/test/monaco"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetMonacoConnector(ctx)

	// One POST-list object and one GET-list object, to cover both read shapes.
	for _, objectName := range []string{"contacts", "accounts", "tags", "users"} {
		if err := testRead(ctx, conn, objectName, time.Time{}); err != nil {
			slog.Error(err.Error())
		}
	}

	// Incremental read. contacts is in incrementalObjects, so this should send
	// an updated_at/greater_than filter rule. Watch for a 4xx here: it is the
	// signal that updated_at is not a filterable field, which we could not
	// verify without credentials.
	if err := testRead(ctx, conn, "contacts", time.Now().AddDate(0, -1, 0)); err != nil {
		slog.Error(err.Error())
	}

	// audiences deliberately ignores Since -- expect a full first page, not an error.
	if err := testRead(ctx, conn, "audiences", time.Now().AddDate(0, -1, 0)); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(ctx context.Context, conn *monaco.Connector, objectName string, since time.Time) error {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id"),
		Since:      since,
		PageSize:   5,
	})
	if err != nil {
		return fmt.Errorf("error reading %s: %w", objectName, err)
	}

	slog.Info("read",
		"object", objectName,
		"since", since,
		"rows", res.Rows,
		"done", res.Done,
		"nextPage", res.NextPage,
	)

	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling %s result: %w", objectName, err)
	}

	slog.Debug(string(jsonStr))

	return nil
}
