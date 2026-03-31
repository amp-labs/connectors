package main

import (
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
	webhookEndpoints = map[string]string{
		"/": "webhooks",
	}
)

func populateWebhooks(
	schemas *staticschema.Metadata[staticschema.FieldMetadataMapV2, any],
	registry datautils.NamedLists[string],
) {
	explorer, err := openapi.WebhooksFileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		nil,
		webhookEndpoints, nil,
		api3.IdenticalObjectLocator,
	)
	goutils.MustBeNil(err)

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		if object.URLPath == "/" {
			object.URLPath = "webhooks"
		}

		displayName := api3.CapitalizeFirstLetterEveryWord(object.ObjectName)

		for _, field := range object.Fields {
			schemas.Add(common.ModuleRoot,
				object.ObjectName, displayName, object.URLPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}
}
