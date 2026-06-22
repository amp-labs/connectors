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
		res, err := conn.Read(ctx, common.ReadParams{
			ObjectName: obj.name,
			Fields:     connectors.Fields(obj.fields...),
		})
		if err != nil {
			utils.Fail("error reading from connector", "object", obj.name, "error", err)
		}

		if res.Rows == 0 {
			slog.Warn(
				"Skipped metadata-vs-read payload check; read returned zero records",
				"object", obj.name,
			)

			continue
		}

		testscenario.ValidateMetadataContainsRead(ctx, conn, obj.name, nil)
	}

	slog.Info("=== Verify metadata exists for all objects ===")
	for _, obj := range objects {
		metadata, err := conn.ListObjectMetadata(ctx, []string{obj.name})
		if err != nil {
			utils.Fail("error listing metadata for connector", "object", obj.name, "error", err)
		}

		objectMeta, ok := metadata.Result[obj.name]
		if !ok || len(objectMeta.Fields) == 0 {
			utils.Fail("metadata schema has no fields", "object", obj.name)
		}

		slog.Info("Metadata defined", "object", obj.name, "fieldCount", len(objectMeta.Fields))
	}

	slog.Info("Breezy read tests completed successfully!")
}
