package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"

	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/getkin/kin-openapi/openapi2"
)

var (
	// assets is a absolute path to the assets file from root of the project.
	assets = "./scripts/openapi/marketo/metadata/asset.json" //nolint:gochecknoglobals
	// leads is a absolute path to the leads file from root of the project.
	leads = "./scripts/openapi/marketo/metadata/mapi.json" //nolint:gochecknoglobals

	// schemas represents the file that holds the generated metadata.
	// Creates it, if not available.
	// This is be created at the root of the project.
	schemas = "schemas.json" //nolint:gochecknoglobals
)

func main() {
	//  read the definitions in the specification file.
	// 5 represents the amount of substrings that will be generated
	// when path of interest is split using `/`
	def, docA, err := constructDefinitions(assets, 5) //nolint:gomnd
	if err != nil {
		panic(err)
	}

	// 4 represents the amount of substrings that will be generated
	// when path of interest is split using `/`
	ldef, docL, err := constructDefinitions(leads, 4) //nolint:gomnd
	if err != nil {
		panic(err)
	}

	// Initializes an empty ObjectMetadata variable
	objectMetadata := make(map[string]scrapper.ObjectMetadata)

	// Add Lead metadata details
	objectMetadata = generateMetadata(ldef, docL, objectMetadata)

	// Adds Assets Metadata details to the same variable declared above.
	objectMetadata = generateMetadata(def, docA, objectMetadata)

	// wrap objectMetadata in `data` to not break the fileManager that reads the schema.
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
	f, err := os.Create(schemas)
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
) map[string]scrapper.ObjectMetadata {
	for obj, dfn := range objDefs {
		schem := doc.Definitions[dfn].Value.Properties

		// Reading the item key that will contain the metadata keys.
		result := schem["result"].Value.Items

		m := cleanDefinitions(result.Ref)
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

	return objectMetadata
}
