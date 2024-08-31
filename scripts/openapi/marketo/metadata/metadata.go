package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
)

func main() {
	//  read the definitions in the specificatios files.
	def, docA, err := constructDefinitions("./scripts/openapi/marketo/metadata/asset.json", 5) //nolint:gomnd
	if err != nil {
		panic(err)
	}

	ldef, docL, err := constructDefinitions("./scripts/openapi/marketo/metadata/mapi.json", 4) //nolint:gomnd
	if err != nil {
		panic(err)
	}

	// Initializes an empty ObjectMetadata variable
	objectMetadata := make(map[string]scrapper.ObjectMetadata)

	// Add Lead metadata details
	objectMetadata, err = generateMetadata(ldef, docL, objectMetadata)
	if err != nil {
		panic(err)
	}

	// Adds Assets Metadata details to the same variable declared above.
	objectMetadata, err = generateMetadata(def, docA, objectMetadata)
	if err != nil {
		panic(err)
	}

	// wrap objectMetadata in `data` to not break the scrapper API.
	data := map[string]any{
		"data": objectMetadata,
	}

	mb, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	// Create a `schemas.json` file and Adds the metadata details to the file.
	if err = writefile(mb); err != nil {
		panic(err)
	}
}

func writefile(b []byte) error {
	f, err := os.Create("schemas.json")
	if err != nil {
		return err
	}

	if _, err := f.Write(b); err != nil {
		return err
	}

	slog.Info("Successfully generated the metadata, written them to schemas.json")

	return nil
}

func constructDefinitions(file string, length int) (map[string]string, openapi2.T, error) {
	definitions := make(map[string]string)

	f, err := os.ReadFile(file)
	if err != nil {
		return nil, openapi2.T{}, err
	}

	var doc openapi2.T
	if err = json.Unmarshal(f, &doc); err != nil {
		return nil, openapi2.T{}, err
	}

	for k, v := range doc.Paths {
		if pathLength(k) == length && v.Get != nil {
			obj := retrieveObject(k)

			for _, j := range v.Get.Responses {
				dfn := cleanDefinitions(j.Schema.Ref)
				definitions[obj] = dfn
			}
		}
	}

	return definitions, doc, nil
}

func pathLength(path string) int {
	return len(strings.Split(path, "/"))
}

func retrieveObject(path string) string {
	s := strings.Split(path, "/")
	sWithJSON := s[len(s)-1]

	return strings.TrimSuffix(sWithJSON, ".json")
}

func cleanDefinitions(def string) string {
	s := strings.Split(def, "/")

	return s[len(s)-1]
}

func generateMetadata(objDefs map[string]string,
	doc openapi2.T, objectMetadata map[string]scrapper.ObjectMetadata,
) (map[string]scrapper.ObjectMetadata, error) {
	for obj, dfn := range objDefs {
		schemas := doc.Definitions[dfn].Value.Properties

		// This is attached to the marketo
		result, err := schemas["result"].Value.JSONLookup("items")
		if err != nil {
			return nil, err
		}

		r, ok := result.(*openapi3.Ref)
		if !ok {
			return nil, fmt.Errorf("failed the assertion, the response data type is not expected") //nolint:goerr113
		}

		m := cleanDefinitions(r.Ref)
		lschems := doc.Definitions[m].Value.Properties

		fields := make(map[string]string)

		for k := range lschems {
			fields[k] = k
		}

		om := scrapper.ObjectMetadata{
			DisplayName: obj,
			FieldsMap:   fields,
		}

		objectMetadata[obj] = om
	}

	return objectMetadata, nil
}
