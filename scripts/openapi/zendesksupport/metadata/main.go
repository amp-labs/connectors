package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/scripts/openapi/zendesksupport/metadata/helpcenter"
	"github.com/amp-labs/connectors/scripts/openapi/zendesksupport/metadata/support"
	"github.com/amp-labs/connectors/tools/scrapper"
)

func main() {
	schemas := staticschema.NewExtendedMetadata[staticschema.FieldMetadataMapV2, metadata.CustomProperties]()
	registry := datautils.NamedLists[string]{}
	lists := datautils.IndexedLists[common.ModuleID, metadatadef.ExtendedSchema[metadata.CustomProperties]]{}

	lists.Add(providers.ModuleZendeskTicketing, support.Objects()...)
	lists.Add(providers.ModuleZendeskHelpCenter, helpcenter.Objects()...)

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
					utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
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
