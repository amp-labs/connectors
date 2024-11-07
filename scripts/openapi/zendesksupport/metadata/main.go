package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/zendesksupport"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
	"github.com/amp-labs/connectors/scripts/openapi/zendesksupport/metadata/helpcenter"
	"github.com/amp-labs/connectors/scripts/openapi/zendesksupport/metadata/support"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

func main() {
	schemas := staticschema.NewMetadata()
	registry := datautils.NamedLists[string]{}
	lists := datautils.IndexedLists[common.ModuleID, api3.Schema]{}

	lists.Add(zendesksupport.ModuleTicketing, support.Objects()...)
	lists.Add(zendesksupport.ModuleHelpCenter, helpcenter.Objects()...)

	for module, objects := range lists {
		for _, object := range objects {
			if object.Problem != nil {
				slog.Error("schema not extracted",
					"objectName", object.ObjectName,
					"error", object.Problem,
				)
			}

			for _, field := range object.Fields {
				schemas.Add(module, object.ObjectName, object.DisplayName, field, object.URLPath, nil)
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
