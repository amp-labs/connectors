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

	name := fmt.Sprintf("Amp WD Project %s", gofakeit.Word())

	createPayload := map[string]any{
		"name":     name,
		"billable": true,
	}

	updatePayload := map[string]any{
		"name": name + " Updated",
	}

	slog.Info("=== time_tracking_projects (create -> update -> delete) ===")
	testscenario.ValidateCreateUpdateDelete(ctx, conn, "time_tracking_projects", createPayload, updatePayload,
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("id", "name"),
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"name": name + " Updated",
			},
		},
	)
}
