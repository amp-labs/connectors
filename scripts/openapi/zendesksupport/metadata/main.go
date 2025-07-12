package main

import (
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/scripts/openapi/zendesksupport/metadata/helpcenter"
	"github.com/amp-labs/connectors/scripts/openapi/zendesksupport/metadata/support"
	"github.com/amp-labs/connectors/tools/scrapper"
)

func main() {
	schemas := staticschema.NewExtendedMetadata[staticschema.FieldMetadataMapV2, metadata.CustomProperties]()
	registry := datautils.NamedLists[string]{}

	objects := append(
		support.Objects(),
		helpcenter.Objects()...,
	)

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		urlPath, _ := strings.CutPrefix(object.URLPath, "/api/v2")

		for _, field := range object.Fields {
			schemas.Add(common.ModuleRoot, object.ObjectName, object.DisplayName, urlPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.FlushSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}
