// Extracts list endpoint schemas from OpenAPI spec and writes providers/okta/metadata/schemas.json.
package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/okta/metadata"
	"github.com/amp-labs/connectors/providers/okta/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	// Only process the endpoints we support.
	allowEndpoints = []string{
		"/api/v1/users",
		"/api/v1/groups",
		"/api/v1/apps",
		"/api/v1/logs",
		"/api/v1/devices",
		"/api/v1/idps",
		"/api/v1/authorizationServers",
		"/api/v1/trustedOrigins",
		"/api/v1/zones",
		"/api/v1/brands",
		"/api/v1/domains",
		"/api/v1/authenticators",
		"/api/v1/policies",
		"/api/v1/eventHooks",
		"/api/v1/features",
	}

	// Display names that differ from the default objectName-based derivation.
	displayNameOverride = map[string]string{
		"apps":                 "Applications",
		"logs":                 "System Log",
		"idps":                 "Identity Providers",
		"authorizationServers": "Authorization Servers",
		"trustedOrigins":       "Trusted Origins",
		"zones":                "Network Zones",
		"eventHooks":           "Event Hooks",
	}

	// Okta returns root-level arrays for most endpoints.
	// The domains endpoint wraps the array in a "domains" key.
	objectNameToResponseField = datautils.NewDefaultMap(map[string]string{
		"domains": "domains",
	},
		func(objectName string) string {
			// Empty string means root-level array response.
			return ""
		},
	)
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)

	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewAllowPathStrategy(allowEndpoints),
		nil, displayNameOverride,
		api3.CustomMappingObjectCheck(objectNameToResponseField),
	)

	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)

			continue
		}

		for _, field := range object.Fields {
			schemas.Add("", object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	// The OpenAPI spec uses allOf/$ref patterns that the tooling cannot auto-extract
	// for devices (DeviceList allOf Device) and policies (bare $ref to Policy).
	// These are added manually based on the Okta API documentation.
	addDevicesObject(schemas)
	addPoliciesObject(schemas)

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func addDevicesObject(schemas *staticschema.Metadata[staticschema.FieldMetadataMapV2, any]) {
	strField := func(name string) staticschema.FieldMetadata {
		return staticschema.FieldMetadata{
			DisplayName: name, ValueType: common.ValueTypeString, ProviderType: "string",
		}
	}

	dateField := func(name string) staticschema.FieldMetadata {
		return staticschema.FieldMetadata{
			DisplayName: name, ValueType: common.ValueTypeString, ProviderType: "date-time",
		}
	}

	objField := func(name string) staticschema.FieldMetadata {
		return staticschema.FieldMetadata{
			DisplayName: name, ValueType: common.ValueTypeOther, ProviderType: "object",
		}
	}

	fields := map[string]staticschema.FieldMetadata{
		"id":                  strField("id"),
		"status":              strField("status"),
		"created":             dateField("created"),
		"lastUpdated":         dateField("lastUpdated"),
		"profile":             objField("profile"),
		"resourceType":        strField("resourceType"),
		"resourceDisplayName": objField("resourceDisplayName"),
		"resourceAlternateId": strField("resourceAlternateId"),
		"resourceId":          strField("resourceId"),
		"_links":              objField("_links"),
	}

	for fieldName, fieldMeta := range fields {
		schemas.Add("", "devices", "Devices", "/api/v1/devices", "",
			staticschema.FieldMetadataMapV2{fieldName: fieldMeta}, nil, nil)
	}
}

func addPoliciesObject(schemas *staticschema.Metadata[staticschema.FieldMetadataMapV2, any]) {
	str := func(name string) staticschema.FieldMetadata {
		return staticschema.FieldMetadata{
			DisplayName: name, ValueType: common.ValueTypeString, ProviderType: "string",
		}
	}

	date := func(name string) staticschema.FieldMetadata {
		return staticschema.FieldMetadata{
			DisplayName: name, ValueType: common.ValueTypeString, ProviderType: "date-time",
		}
	}

	obj := func(name string) staticschema.FieldMetadata {
		return staticschema.FieldMetadata{
			DisplayName: name, ValueType: common.ValueTypeOther, ProviderType: "object",
		}
	}

	fields := map[string]staticschema.FieldMetadata{
		"id":          str("id"),
		"name":        str("name"),
		"type":        str("type"),
		"status":      str("status"),
		"description": str("description"),
		"priority": {
			DisplayName:  "priority",
			ValueType:    common.ValueTypeInt,
			ProviderType: "integer",
		},
		"system": {
			DisplayName:  "system",
			ValueType:    common.ValueTypeBoolean,
			ProviderType: "boolean",
		},
		"created":     date("created"),
		"lastUpdated": date("lastUpdated"),
		"conditions":  obj("conditions"),
		"_links":      obj("_links"),
	}

	for fieldName, fieldMeta := range fields {
		schemas.Add("", "policies", "Policies", "/api/v1/policies", "",
			staticschema.FieldMetadataMapV2{fieldName: fieldMeta}, nil, nil)
	}
}
