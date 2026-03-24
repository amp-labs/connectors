package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/bentley/metadata"
	"github.com/amp-labs/connectors/tools/scrapper"
)

func main() {
	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	populateITwins(schemas, registry)
	populateCesium(schemas, registry)
	populateLibrary(schemas, registry)
	populateEdfs(schemas, registry)
	populateContextCapture(schemas, registry)
	populateRealityManagement(schemas, registry)
	populateRealityAnalysis(schemas, registry)
	populateRealityConversion(schemas, registry)
	populateWebhooks(schemas, registry)

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}
