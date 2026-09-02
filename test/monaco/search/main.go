package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

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

	// The spec's own example: look a contact up by exact email. This is the
	// case most likely to be accepted, since `equals` on email is the one
	// equality example Monaco documents.
	if err := testSearch(ctx, conn, "contacts", "email", "jane@acme.com"); err != nil {
		slog.Error(err.Error())
	}

	// A UUID-valued field. Watch for a 4xx here specifically: Monaco's examples
	// use `equals` for sequences.contact_id but `is` for meetings.account_id,
	// and which operators a field accepts is only discoverable from
	// GET /v1/schemas/{entity}, which we cannot call without a key. A failure
	// here means the eq -> equals mapping needs to become per-field.
	if err := testSearch(ctx, conn, "meetings", "account_id",
		"550e8400-e29b-41d4-a716-446655440000"); err != nil {
		slog.Error(err.Error())
	}

	// Not searchable: its list request carries no filters at all.
	if err := testSearch(ctx, conn, "audiences", "name", "Q3 targets"); err != nil {
		slog.Error(err.Error())
	}
}

func testSearch(ctx context.Context, conn *monaco.Connector, objectName, field string, value any) error {
	res, err := conn.Search(ctx, &common.SearchParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id"),
		Filter: common.SearchFilter{
			FieldFilters: []common.FieldFilter{{
				FieldName: field,
				Operator:  common.FilterOperatorEQ,
				Value:     value,
			}},
		},
		Limit: 5,
	})
	if err != nil {
		return fmt.Errorf("error searching %s on %s: %w", objectName, field, err)
	}

	slog.Info("search",
		"object", objectName,
		"filter", fmt.Sprintf("%s eq %v", field, value),
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
