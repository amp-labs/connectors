package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers/stripe/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer()
	goutils.MustBeNil(err)

	utilsopenapi.PrintWriteEndpointsSummary(utilsopenapi.WriteExplorerArgs{
		Explorer: explorer,
	})

	slog.Info("Completed.")
}
