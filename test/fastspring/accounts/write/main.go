package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/fastspring"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetFastSpringConnector(ctx)

	email := fmt.Sprintf("amp.integration.%s@example.com", gofakeit.UUID())
	slog.Info("=== Creating account (no delete API) ===", "email", email)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "accounts",
		RecordData: map[string]any{
			"contact": map[string]any{
				"first": gofakeit.FirstName(),
				"last":  gofakeit.LastName(),
				"email": email,
			},
		},
	})
	if err != nil {
		slog.Error("Failed to create account", "error", err)
		os.Exit(1)
	}

	utils.DumpJSON(res, os.Stdout)
	slog.Info("Account write completed; remove test data in FastSpring dashboard if needed", "recordId", res.RecordId)
}
