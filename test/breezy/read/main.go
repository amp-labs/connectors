package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/breezy"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetBreezyConnector(ctx)

	objects := []struct {
		name   string
		fields []string
	}{
		{name: "companies", fields: []string{"_id", "name"}},
		{name: "positions", fields: []string{"_id", "name", "state"}},
		{name: "pipelines", fields: []string{"_id", "name"}},
		{name: "categories", fields: []string{"id", "name"}},
		{name: "departments", fields: []string{"_id", "name"}},
		{name: "questionnaires", fields: []string{"_id", "name"}},
		{name: "templates", fields: []string{"_id", "name"}},
		{name: "webhook_endpoints", fields: []string{"id", "url", "status"}},
	}

	for _, obj := range objects {
		slog.Info("=== Read " + obj.name + " ===")
		testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
			ObjectName: obj.name,
			Fields:     connectors.Fields(obj.fields...),
		})
	}

	slog.Info("=== Metadata vs Read validation ===")
	for _, obj := range objects {
		testscenario.ValidateMetadataContainsRead(ctx, conn, obj.name, nil)
	}

	slog.Info("Breezy read tests completed successfully!")
}
