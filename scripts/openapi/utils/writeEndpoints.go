package utilsopenapi

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

type WriteExplorerArgs struct {
	// Name is useful when more than one explorer is passed into PrintWriteEndpointsSummary.
	Name        string
	Explorer    *api3.Explorer
	PathMatcher api3.PathMatcher
}

func PrintWriteEndpointsSummary(args ...WriteExplorerArgs) {
	inputs := make(datautils.Map[string, WriteExplorerArgs])

	for _, arg := range args {
		if arg.PathMatcher == nil {
			arg.PathMatcher = api3.DefaultPathMatcher{}
		}

		inputs[arg.Name] = arg
	}

	names := inputs.Keys()
	sort.Strings(names)

	for _, name := range names {
		fmt.Println("=====================") // nolint:forbidigo
		fmt.Printf("OpenAPI %v\n", name)     // nolint:forbidigo
		fmt.Println("=====================") // nolint:forbidigo

		input := inputs[name]
		endpoints, err := input.Explorer.GetEndpointOperations(
			input.PathMatcher, http.MethodPost, http.MethodPatch, http.MethodPut,
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
}
