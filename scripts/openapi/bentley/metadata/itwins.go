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
	iTwinsIgnoreEndpoints = []string{
		"/myprimaryaccount",
	}

	// The explorer derives object names from the URL path after the server base.
	// The server base is already "https://api.bentley.com/itwins", so the root
	// endpoint "/" has no path segment left — the object name comes out as "".
	// We map "" to "iTwins" so the schema builder knows which response key to use.
	iTwinsObjectNameToResponseField = datautils.NewDefaultMap(map[string]string{
		"favorites": "iTwins",
		"recents":   "iTwins",
		"":          "iTwins",
	},
		func(objectName string) (fieldName string) {
			return objectName
		})
)

//nolint:dupl
func populateITwins(
	schemas *staticschema.Metadata[staticschema.FieldMetadataMapV2, any],
	registry datautils.NamedLists[string],
) {
	explorer, err := openapi.ITwinsFileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(iTwinsIgnoreEndpoints),
		nil, nil,
		api3.CustomMappingObjectCheck(iTwinsObjectNameToResponseField),
	)
	goutils.MustBeNil(err)

	// Bentley has many APIs all sharing one schemas.json. We use a prefix so
	// objects from different files don't clash, e.g. "itwins/favorites" vs
	// "library/manufacturers".
	prefix := "itwins"

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
