//nolint:dupl
package main

import (
	"fmt"
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/bentley/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

// nolint:gochecknoglobals
var (
	realityConversionIgnoreEndpoints = []string{
		"/myprimaryaccount",
	}

	realityConversionObjectNameToResponseField = datautils.NewDefaultMap(map[string]string{
		"realityConversion": "items",
	},
		func(objectName string) (fieldName string) {
			return objectName
		})
)

//nolint:dupl
func populateRealityConversion(
	schemas *staticschema.Metadata[staticschema.FieldMetadataMapV2, any],
	registry datautils.NamedLists[string],
) {
	explorer, err := openapi.RealityConversionFileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(realityConversionIgnoreEndpoints),
		nil, nil,
		api3.CustomMappingObjectCheck(realityConversionObjectNameToResponseField),
	)
	goutils.MustBeNil(err)

	prefix := "realityconversion"

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		objectName := fmt.Sprintf("%s/%s", prefix, object.ObjectName)
		displayName := api3.CapitalizeFirstLetterEveryWord(prefix) + " " +
			api3.CapitalizeFirstLetterEveryWord(object.ObjectName)

		for _, field := range object.Fields {
			schemas.Add(common.ModuleRoot,
				objectName, displayName, objectName, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}
}
