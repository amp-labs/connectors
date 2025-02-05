package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"sort"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers/keap/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

func main() {
	exp1, err := openapi.Version1FileManager.GetExplorer()
	goutils.MustBeNil(err)
	exp2, err := openapi.Version2FileManager.GetExplorer()
	goutils.MustBeNil(err)

	explorers := datautils.Map[string, *api3.Explorer]{"file1": exp1, "file2": exp2}
	names := explorers.Keys()
	sort.Strings(names)

	for _, name := range names {
		fmt.Println("=====================") // nolint:forbidigo
		fmt.Printf("OpenAPI %v\n", name)     // nolint:forbidigo
		fmt.Println("=====================") // nolint:forbidigo

		endpoints, err := explorers[name].GetEndpointOperations(
			api3.DefaultPathMatcher{}, http.MethodPost, http.MethodPatch, http.MethodPut,
		)
		goutils.MustBeNil(err)

		displays := make([]string, len(endpoints))
		for i, endpoint := range endpoints {
			displays[i] = endpoint.String()
		}

		sort.Strings(displays)

		for _, display := range displays {
			fmt.Println(display) // nolint:forbidigo
		}

		fmt.Printf("\nTotal number of endpoints %v\n\n", len(endpoints)) // nolint:forbidigo
	}

	slog.Info("Completed.")
}
