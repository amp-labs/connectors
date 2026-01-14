package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/marketo/metadata"
	"github.com/getkin/kin-openapi/openapi2"
)

var (
	// assets is a absolute path to the assets file from root of the project.
	assets = "./scripts/openapi/marketo/metadata/asset.json" //nolint:gochecknoglobals
	// leads is a absolute path to the leads file from root of the project.
	leads = "./scripts/openapi/marketo/metadata/mapi.json" //nolint:gochecknoglobals
)

func main() {
	//  read the definitions in the specification file.
	// 5 represents the amount of substrings that will be generated
	// when path of interest is split using `/`
	def, docA, err := constructDefinitions(assets, 5) //nolint:mnd
	goutils.MustBeNil(err)

	// 4 represents the amount of substrings that will be generated
	// when path of interest is split using `/`
	ldef, docL, err := constructDefinitions(leads, 4) //nolint:mnd
	goutils.MustBeNil(err)

	// Initializes an empty map of leads ObjectMetadata
	leadsMetadata := make(map[string]staticschema.Object[staticschema.FieldMetadataMapV1, any])

	// Initializes an empty map of assets ObjectMetadata
	assetsMetadata := make(map[string]staticschema.Object[staticschema.FieldMetadataMapV1, any])

	merged := make(map[string]staticschema.Object[staticschema.FieldMetadataMapV1, any])

	// Adds Leads metadata details
	leadsMetadata = generateMetadata(ldef, docL, leadsMetadata)

	// Adds Assets Metadata details
	assetsMetadata = generateMetadata(def, docA, assetsMetadata)

	maps.Copy(merged, leadsMetadata)
	maps.Copy(merged, assetsMetadata)

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(&staticschema.Metadata[staticschema.FieldMetadataMapV1, any]{
		Modules: map[common.ModuleID]staticschema.Module[staticschema.FieldMetadataMapV1, any]{
			common.ModuleRoot: {
				Objects: merged,
			},
		},
	}))
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
	doc openapi2.T, objectMetadata map[string]staticschema.Object[staticschema.FieldMetadataMapV1, any],
) map[string]staticschema.Object[staticschema.FieldMetadataMapV1, any] {
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

		om := staticschema.Object[staticschema.FieldMetadataMapV1, any]{
			DisplayName: obj,
			URLPath:     fmt.Sprintf("/%v", obj),
			Fields:      fields,
		}

		objectMetadata[obj] = om
	}

	return objectMetadata
}
