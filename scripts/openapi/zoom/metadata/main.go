package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/zoom"
	"github.com/amp-labs/connectors/providers/zoom/metadata"
	"github.com/amp-labs/connectors/scripts/openapi/zoom/metadata/meeting"
	"github.com/amp-labs/connectors/scripts/openapi/zoom/metadata/user"
	"github.com/amp-labs/connectors/tools/scrapper"
)

func main() {
	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV1]()
	registry := datautils.NamedLists[string]{}
	lists := datautils.IndexedLists[common.ModuleID, metadatadef.Schema]{}

	lists.Add(zoom.ModuleUser, user.Objects()...)
	lists.Add(zoom.ModuleMeeting, meeting.Objects()...)

	for module, objects := range lists {
		for _, object := range objects {
			if object.Problem != nil {
				slog.Error("schema not extracted",
					"objectName", object.ObjectName,
					"error", object.Problem,
				)
			}

			for _, field := range object.Fields {
				schemas.Add(module, object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
					staticschema.FieldMetadataMapV1{
						field.Name: field.Name,
					}, nil, nil)
			}

			for _, queryParam := range object.QueryParams {
				registry.Add(queryParam, object.ObjectName)
			}
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}
