package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/scripts/openapi/teamwork/internal/files"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
)

func main() {
	explorer, err := files.InputTeamwork.GetExplorer()
	goutils.MustBeNil(err)

	utilsopenapi.PrintWriteEndpointsSummary(utilsopenapi.WriteExplorerArgs[any]{
		Explorer: explorer,
	})

	slog.Info("Completed.")
}
