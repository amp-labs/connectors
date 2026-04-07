package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/bentley"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := bentley.GetBentleyConnector(ctx)

	slog.Info("Reading itwins")

	res, err := connector.Read(ctx, common.ReadParams{
		ObjectName: "itwins",
		Fields:     datautils.NewStringSet("class", "id", "displayName"),
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	slog.Info("Reading curated-content/cesium")

	res, err = connector.Read(ctx, common.ReadParams{
		ObjectName: "curated-content/cesium",
		Fields:     datautils.NewStringSet("id", "status", "attribution"),
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}
