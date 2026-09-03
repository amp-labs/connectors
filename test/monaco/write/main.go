package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

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

	// Create then update a tag: the cheapest round trip that exercises both
	// halves of Write, and the least intrusive object to leave behind.
	created, err := testWrite(ctx, conn, common.WriteParams{
		ObjectName: "tags",
		RecordData: map[string]any{
			"object": "contact",
			"name":   "ampersand-smoke-test",
			"color":  "#22c55e",
		},
	})
	if err != nil {
		slog.Error(err.Error())

		return
	}

	if _, err := testWrite(ctx, conn, common.WriteParams{
		ObjectName: "tags",
		RecordId:   created.RecordId,
		RecordData: map[string]any{"name": "ampersand-smoke-test-renamed"},
	}); err != nil {
		slog.Error(err.Error())
	}

	// Accounts have no create endpoint, so this must go out as PUT (upsert).
	// `domain` is the required key it matches on. Expect an update rather than
	// an insert if the domain already exists.
	if _, err := testWrite(ctx, conn, common.WriteParams{
		ObjectName: "accounts",
		RecordData: map[string]any{
			"domain": "ampersand-smoke-test.example.com",
			"name":   "Ampersand Smoke Test",
		},
	}); err != nil {
		slog.Error(err.Error())
	}
}

func testWrite(
	ctx context.Context, conn *monaco.Connector, params common.WriteParams,
) (*common.WriteResult, error) {
	res, err := conn.Write(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error writing %s: %w", params.ObjectName, err)
	}

	slog.Info("write",
		"object", params.ObjectName,
		"create", params.RecordId == "",
		"success", res.Success,
		"recordId", res.RecordId,
	)

	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshalling %s result: %w", params.ObjectName, err)
	}

	slog.Debug(string(jsonStr))

	return res, nil
}
