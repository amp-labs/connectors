package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers/stripe/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer()
	goutils.MustBeNil(err)

	endpoints, err := explorer.GetEndpointOperations(
		api3.DefaultPathMatcher{}, http.MethodPost, http.MethodPatch, http.MethodPut,
	)
	goutils.MustBeNil(err)

	for _, endpoint := range endpoints {
		fmt.Println(endpoint) // nolint:forbidigo
	}

	fmt.Println("Total number of endpoints", len(endpoints)) // nolint:forbidigo

	slog.Info("Completed.")
}
