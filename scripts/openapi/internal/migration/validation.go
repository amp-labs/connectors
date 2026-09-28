// nolint:forbidigo
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/amp-labs/connectors/internal/fileregistry"
	"github.com/google/go-cmp/cmp"
)

type Object struct {
	DisplayName string                    `json:"displayName"`
	Fields      map[string]map[string]any `json:"fields"`
	Path        string                    `json:"path"`
	ResponseKey string                    `json:"responseKey"`
	ReferenceTo []string                  `json:"referenceTo"`
}

func main() {
	oldDataFile := flag.String("old", "", "old file: schemas.json")
	newDataFile := flag.String("new", "", "new file: objectsMetadata.json")
	newEndpointsFile := flag.String("endpoints", "", "file: endpoints.json")
	moduleName := flag.String("module", "", "module name in old schemas.json")
	flag.Parse()

	if *oldDataFile == "" || *newDataFile == "" || *newEndpointsFile == "" || *moduleName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	oldData, err := os.ReadFile(*oldDataFile)
	must(err)
	newData, err := os.ReadFile(*newDataFile)
	must(err)
	newEndpoints, err := os.ReadFile(*newEndpointsFile)
	must(err)

	oldObjs, err := extractObjects(oldData, nil, *moduleName)
	must(err)
	newObjs, err := extractObjects(newData, newEndpoints, *moduleName)
	must(err)

	if diff := cmp.Diff(oldObjs, newObjs, cmp.AllowUnexported(Object{})); diff != "" {
		fmt.Println("different:")
		fmt.Println(diff)

		return
	}

	fmt.Println("equal")
}

func extractObjects(data []byte, endpointsData []byte, moduleName string) (map[string]Object, error) {
	output := make(map[string]Object)

	if endpointsData == nil {
		document := fileregistry.MustParseJSON[OldFormat](data)

		for name, object := range document.Modules[moduleName].Objects {
			output[name] = Object{
				DisplayName: object.DisplayName,
				Fields:      object.Fields,
				Path:        object.Path,
				ResponseKey: object.ResponseKey,
				ReferenceTo: object.ReferenceTo,
			}
		}

		return output, nil
	}

	document := fileregistry.MustParseJSON[NewFormat](data)
	endpoints := fileregistry.MustParseJSON[NewEndpoints](endpointsData)

	for name, object := range document {
		endpoint, ok := endpoints[name]
		if !ok {
			must(fmt.Errorf("endpoints.json does not have object with name %v", name)) // nolint:err113
		}

		if endpoint.Read.Url != object.Path {
			must(fmt.Errorf("endpoints.json and objectSchema.json "+ // nolint:err113
				"do not agree on the object URL path: object(%v): %v =/= %v", name, endpoint.Read.Url, object.Path))
		}

		output[name] = Object{
			DisplayName: object.DisplayName,
			Fields:      object.Fields,
			Path:        object.Path,
			ResponseKey: endpoint.Read.ResponseKey,
			ReferenceTo: object.ReferenceTo,
		}
	}

	return output, nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type OldFormat struct {
	Modules map[string]Module `json:"modules"`
}

type Module struct {
	Objects map[string]struct {
		DisplayName string                    `json:"displayName"`
		Path        string                    `json:"path"`
		ResponseKey string                    `json:"responseKey"`
		Fields      map[string]map[string]any `json:"fields"`
		ReferenceTo []string                  `json:"referenceTo"`
	} `json:"objects"`
}

type NewFormat map[string]struct {
	Path        string                    `json:"path"`
	DisplayName string                    `json:"displayName"`
	Fields      map[string]map[string]any `json:"fields"`
	ReferenceTo []string                  `json:"referenceTo"`
}

type NewEndpoints map[string]struct {
	Read struct {
		Url         string `json:"url"`
		Operation   string `json:"operation"`
		ResponseKey string `json:"responseKey"`
	} `json:"read"`
}
