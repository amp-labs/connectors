package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/ramp"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetRampConnector(ctx)

	slog.Info("> TEST Create department")

	createResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "departments",
		RecordData: map[string]any{
			"name": "Test Department (connector test)",
		},
	})
	if err != nil {
		utils.Fail("error creating department", "error", err)
	}

	printJSON(createResult)

	if createResult.RecordId == "" {
		utils.Fail("expected a record ID after create")
	}

	slog.Info("> TEST Update department", "id", createResult.RecordId)

	updateResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "departments",
		RecordId:   createResult.RecordId,
		RecordData: map[string]any{
			"name": "Test Department (updated)",
		},
	})
	if err != nil {
		utils.Fail("error updating department", "error", err)
	}

	printJSON(updateResult)

	slog.Info("> TEST Create location")

	locResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "locations",
		RecordData: map[string]any{
			"name": "Test Location (connector test)",
		},
	})
	if err != nil {
		utils.Fail("error creating location", "error", err)
	}

	printJSON(locResult)

	slog.Info("Done")
}

func printJSON(v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		utils.Fail("json marshal error", "error", err)
	}

	fmt.Println(string(b))
}
