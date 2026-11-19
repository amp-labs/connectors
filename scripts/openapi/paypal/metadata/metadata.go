package main

import (
	_ "embed"
	"log"
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/staticschema"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

/*
The PayPal OpeAPI files are broken down into their APIs.
Their located at: https://github.com/paypal/paypal-rest-api-specifications/tree/main/openapi.
This example just shows how one OpenAPI schemas can be calculated. The schemas file in the connector
is an aggregated file, after running all the OpenAPI files.
*/

var (
	//go:embed reporting.json
	apiFile []byte

	//go:embed schemas.json
	schemas []byte

	SchemaFileManager = scrapper.NewMetadataFileManager[staticschema.FieldMetadataMapV2]( // nolint:gochecknoglobals
		schemas, fileconv.NewSiblingFileLocator())

	// Schemas is cached Object schemas.
	Schemas = SchemaFileManager.MustLoadSchemas() // nolint:gochecknoglobals

	FileManager = api3.NewOpenapiFileManager[any](apiFile) // nolint:gochecknoglobals
)

func locator(objectName string, fieldName string) bool {
	switch objectName {
	case "disputes", "invoices":
		return fieldName == "items"
	case "webhooks-lookup":
		return fieldName == "links"
	case "webhooks-event-types":
		return fieldName == "event_types"
	case "webhooks-events":
		return fieldName == "events"
	case "web-profiles":
		return fieldName == ""
	case "transactions":
		return fieldName == "transaction_details"
	case "balances":
		return fieldName == "balances"
	}

	return fieldName == objectName
}

func main() {
	explorer, err := FileManager.GetExplorer()
	if err != nil {
		log.Fatalln(err)
	}

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(nil),
		nil, nil,
		locator)
	if err != nil {
		log.Fatalln(err)
	}

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add("", object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	if err := SchemaFileManager.SaveSchemas(schemas); err != nil {
		log.Fatalln(err)
	}

	if err := SchemaFileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)); err != nil {
		log.Fatalln(err)
	}

	slog.Info("Completed.")
}
