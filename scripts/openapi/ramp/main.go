package main

import (
	_ "embed"
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/ramp/metadata"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

//go:embed developer-api.json
var apiFile []byte // nolint:gochecknoglobals

var (
	// objectEndpoints maps URL paths to object names only for multi-segment paths
	// where the last segment alone is ambiguous. Object names are the full path
	// after /developer/v1/, preserving hyphens and slashes as-is.
	objectEndpoints = map[string]string{ // nolint:gochecknoglobals
		"/developer/v1/bills/drafts":        "bills/drafts",
		"/developer/v1/audit-logs/events":   "audit-logs/events",
		"/developer/v1/accounting/accounts": "accounting/accounts",
		"/developer/v1/accounting/vendors":  "accounting/vendors",
		"/developer/v1/accounting/fields":   "accounting/fields",
	}

	// displayNameOverride supplies human-readable names for object names that
	// contain slashes, which the default title-case processor cannot handle.
	displayNameOverride = map[string]string{ // nolint:gochecknoglobals
		"bills/drafts":        "Bill Drafts",
		"accounting/accounts": "Accounting Accounts",
		"accounting/vendors":  "Accounting Vendors",
		"accounting/fields":   "Accounting Fields",
	}

	// allowPaths is the exhaustive list of Ramp API endpoints included in this connector.
	// audit_logs and trips are excluded here because their OpenAPI spec has a doubly-nested
	// array bug; they are added manually below.
	allowPaths = []string{ // nolint:gochecknoglobals
		"/developer/v1/transactions",
		"/developer/v1/users",
		"/developer/v1/cards",
		"/developer/v1/departments",
		"/developer/v1/locations",
		"/developer/v1/vendors",
		"/developer/v1/limits",
		"/developer/v1/reimbursements",
		"/developer/v1/bills",
		"/developer/v1/bills/drafts",
		"/developer/v1/receipts",
		"/developer/v1/spend-programs",
		"/developer/v1/statements",
		"/developer/v1/transfers",
		"/developer/v1/cashbacks",
		"/developer/v1/merchants",
		"/developer/v1/entities",
		"/developer/v1/bank-accounts",
		"/developer/v1/purchase-orders",
		"/developer/v1/item-receipts",
		"/developer/v1/repayments",
		"/developer/v1/memos",
		"/developer/v1/unified-requests",
		"/developer/v1/accounting/accounts",
		"/developer/v1/accounting/vendors",
		"/developer/v1/accounting/fields",
	}
)

func main() {
	fileManager := api3.NewOpenapiFileManager[any](apiFile)

	explorer, err := fileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)

	goutils.MustBeNil(err)

	readObjects, err := explorer.ReadObjectsGet(
		api3.NewAllowPathStrategy(allowPaths),
		objectEndpoints,
		displayNameOverride,
		// All Ramp paginated list responses return items under the "data" key.
		// unified-requests is a direct array response; DataObjectLocator is not
		// consulted for array-typed schemas, so it is safe to use here universally.
		api3.DataObjectLocator,
	)

	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range readObjects {
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

	// audit-logs/events and trips have a doubly-nested array in the Ramp OpenAPI spec
	// (data: array of array of Resource), which the explorer cannot process.
	// Their fields are added manually from the component schemas.
	for fieldName, field := range auditLogFields {
		schemas.Add("", "audit-logs/events", "Audit Logs", "/developer/v1/audit-logs/events", "data",
			staticschema.FieldMetadataMapV2{fieldName: field}, nil, nil)
	}

	for fieldName, field := range tripFields {
		schemas.Add("", "trips", "Trips", "/developer/v1/trips", "data",
			staticschema.FieldMetadataMapV2{fieldName: field}, nil, nil)
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

// auditLogFields holds fields for audit-logs/events (AuditLogEventResource component).
// Manually defined because the Ramp OpenAPI spec uses a doubly-nested array for data.
var auditLogFields = staticschema.FieldMetadataMapV2{ // nolint:gochecknoglobals
	"actor_details":      {DisplayName: "Actor Details", ValueType: common.ValueTypeOther, ProviderType: "object"},
	"actor_id":           {DisplayName: "Actor Id", ValueType: common.ValueTypeString, ProviderType: "string"},
	"actor_type":         {DisplayName: "Actor Type", ValueType: common.ValueTypeString, ProviderType: "string"},
	"additional_details": {DisplayName: "Additional Details", ValueType: common.ValueTypeString, ProviderType: "string"},
	"event_details":      {DisplayName: "Event Details", ValueType: common.ValueTypeOther, ProviderType: "object"},
	"event_time":         {DisplayName: "Event Time", ValueType: common.ValueTypeString, ProviderType: "string"},
	"event_type":         {DisplayName: "Event Type", ValueType: common.ValueTypeString, ProviderType: "string"},
	"id":                 {DisplayName: "Id", ValueType: common.ValueTypeString, ProviderType: "string"},
	"primary_reference":  {DisplayName: "Primary Reference", ValueType: common.ValueTypeOther, ProviderType: "object"},
	"user_details":       {DisplayName: "User Details", ValueType: common.ValueTypeOther, ProviderType: "object"},
}

// tripFields holds fields for trips (TripResource component).
// Manually defined because the Ramp OpenAPI spec uses a doubly-nested array for data.
var tripFields = staticschema.FieldMetadataMapV2{ // nolint:gochecknoglobals
	"created_at":     {DisplayName: "Created At", ValueType: common.ValueTypeString, ProviderType: "string"},
	"description":    {DisplayName: "Description", ValueType: common.ValueTypeString, ProviderType: "string"},
	"end_date":       {DisplayName: "End Date", ValueType: common.ValueTypeString, ProviderType: "string"},
	"id":             {DisplayName: "Id", ValueType: common.ValueTypeString, ProviderType: "string"},
	"length_in_days": {DisplayName: "Length In Days", ValueType: common.ValueTypeInt, ProviderType: "integer"},
	"locations":      {DisplayName: "Locations", ValueType: common.ValueTypeOther, ProviderType: "array"},
	"name":           {DisplayName: "Name", ValueType: common.ValueTypeString, ProviderType: "string"},
	"spend_events":   {DisplayName: "Spend Events", ValueType: common.ValueTypeOther, ProviderType: "array"},
	"start_date":     {DisplayName: "Start Date", ValueType: common.ValueTypeString, ProviderType: "string"},
	"status":         {DisplayName: "Status", ValueType: common.ValueTypeString, ProviderType: "string"},
	"total_spend":    {DisplayName: "Total Spend", ValueType: common.ValueTypeFloat, ProviderType: "number"},
	"travel_types":   {DisplayName: "Travel Types", ValueType: common.ValueTypeOther, ProviderType: "array"},
	"updated_at":     {DisplayName: "Updated At", ValueType: common.ValueTypeString, ProviderType: "string"},
	"user_id":        {DisplayName: "User Id", ValueType: common.ValueTypeString, ProviderType: "string"},
}
