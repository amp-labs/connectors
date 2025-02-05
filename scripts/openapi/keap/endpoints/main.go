package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers/keap/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
)

func main() {
	exp1, err := openapi.Version1FileManager.GetExplorer()
	goutils.MustBeNil(err)
	exp2, err := openapi.Version2FileManager.GetExplorer()
	goutils.MustBeNil(err)

	utilsopenapi.PrintWriteEndpointsSummary([]utilsopenapi.WriteExplorerArgs{{
		Name:     "file1",
		Explorer: exp1,
	}, {
		Name:     "file2",
		Explorer: exp2,
	}}...)

	slog.Info("Completed.")
}
