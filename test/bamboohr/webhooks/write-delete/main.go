package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/bamboohr"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetBambooHRConnector(ctx)

	name := fmt.Sprintf("Amp WD Webhook %s", gofakeit.Word())
	url := "https://example.com/ampersand-webhook-test"

	createPayload := map[string]any{
		"name":   name,
		"url":    url,
		"format": "json",
		"events": []string{"employee.created"},
	}

	updatePayload := map[string]any{
		"name":   name + " Updated",
		"url":    url,
		"format": "json",
		"events": []string{"employee.created"},
	}

	slog.Info("=== webhooks (create -> update -> delete) ===")
	testscenario.ValidateCreateUpdateDelete(ctx, conn, "webhooks", createPayload, updatePayload,
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("id", "name", "url"),
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"name": name + " Updated",
			},
		},
	)
}
