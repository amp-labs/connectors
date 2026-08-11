// Generates providers/acculynx/metadata/schemas.json from the AccuLynx V2 OpenAPI spec.
//
// Run from repo root:
//
//	go run ./scripts/openapi/acculynx/metadata
package main

import (
	"log/slog"
	"maps"
	"sort"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/acculynx/metadata"
	"github.com/amp-labs/connectors/providers/acculynx/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/getkin/kin-openapi/openapi3"
)

// objectEndpoints maps URL path → ObjectName. Templated paths ({jobId},
// {contactId}, etc.) are kept verbatim; the read connector substitutes them at
// request time via parent-list fan-out.
//
//nolint:gochecknoglobals
var objectEndpoints = map[string]string{
	"/acculynx/countries":        "acculynx/countries",
	"/acculynx/units-of-measure": "acculynx/units-of-measure",
	"/calendars":                 "calendars",
	"/contacts":                  "contacts",
	"/contacts/contact-types":    "contacts/contact-types",
	"/estimates":                 "estimates",
	"/jobs":                      "jobs",
	"/supplements":               "supplements",
	"/users":                     "users",

	"/company-settings/custom-fields":                         "company-settings/custom-fields",
	"/company-settings/job-file-settings/document-folders":    "company-settings/job-file-settings/document-folders",
	"/company-settings/job-file-settings/insurance-companies": "company-settings/job-file-settings/insurance-companies",
	"/company-settings/job-file-settings/job-categories":      "company-settings/job-file-settings/job-categories",
	"/company-settings/job-file-settings/photo-video-tags":    "company-settings/job-file-settings/photo-video-tags",
	"/company-settings/job-file-settings/trade-types":         "company-settings/job-file-settings/trade-types",
	"/company-settings/job-file-settings/work-types":          "company-settings/job-file-settings/work-types",
	"/company-settings/job-file-settings/workflow-milestones": "company-settings/job-file-settings/workflow-milestones",
	"/company-settings/leads/lead-sources":                    "company-settings/leads/lead-sources",
	"/company-settings/location-settings/account-types":       "company-settings/location-settings/account-types",

	// /jobs/{jobId}/{documents,messages,photos-videos,payments/*} are excluded:
	// documents/messages/photos-videos are POST-only in the spec, and the
	// payments family is a singleton aggregate or write-only sub-endpoints.
	"/jobs/{jobId}/contacts":          "jobs/contacts",
	"/jobs/{jobId}/custom-fields":     "jobs/custom-fields",
	"/jobs/{jobId}/estimates":         "jobs/estimates",
	"/jobs/{jobId}/history":           "jobs/history",
	"/jobs/{jobId}/invoices":          "jobs/invoices",
	"/jobs/{jobId}/milestone-history": "jobs/milestone-history",
	"/jobs/{jobId}/representatives":   "jobs/representatives",

	"/contacts/{contactId}/custom-fields":   "contacts/custom-fields",
	"/contacts/{contactId}/email-addresses": "contacts/email-addresses",
	"/contacts/{contactId}/phone-numbers":   "contacts/phone-numbers",

	// /estimates/{estimateId}/sections/{sectionId}/items deferred: 2-level fan-out.
	"/estimates/{estimateId}/sections": "estimates/sections",

	"/supplements/{supplementId}/items":     "supplements/items",
	"/supplements/{supplementId}/notations": "supplements/notations",

	// Read layer must default startDate/endDate; the endpoint requires them.
	"/calendars/{calendarId}/appointments": "calendars/appointments",
}

//nolint:gochecknoglobals
var displayNameOverride = map[string]string{
	"acculynx/countries":        "AccuLynx Countries",
	"acculynx/units-of-measure": "AccuLynx Units of Measure",
	"calendars/appointments":    "Calendar Appointments",
	"contacts/contact-types":    "Contact Types",
	"contacts/custom-fields":    "Contact Custom Fields",
	"contacts/email-addresses":  "Contact Email Addresses",
	"contacts/phone-numbers":    "Contact Phone Numbers",
	"estimates/sections":        "Estimate Sections",
	"jobs/contacts":             "Job Contacts",
	"jobs/custom-fields":        "Job Custom Fields",
	"jobs/estimates":            "Job Estimates",
	"jobs/history":              "Job History",
	"jobs/invoices":             "Job Invoices",
	"jobs/milestone-history":    "Job Milestone History",
	"jobs/representatives":      "Job Representatives",
	"supplements/items":         "Supplement Items",
	"supplements/notations":     "Supplement Notations",

	"company-settings/custom-fields":                         "Company Custom Fields",
	"company-settings/job-file-settings/document-folders":    "Document Folders",
	"company-settings/job-file-settings/insurance-companies": "Insurance Companies",
	"company-settings/job-file-settings/job-categories":      "Job Categories",
	"company-settings/job-file-settings/photo-video-tags":    "Photo and Video Tags",
	"company-settings/job-file-settings/trade-types":         "Trade Types",
	"company-settings/job-file-settings/work-types":          "Work Types",
	"company-settings/job-file-settings/workflow-milestones": "Workflow Milestones",
	"company-settings/leads/lead-sources":                    "Lead Sources",
	"company-settings/location-settings/account-types":       "Account Types",
}

//nolint:gochecknoglobals
var allowedPaths = func() []string {
	paths := make([]string, 0, len(objectEndpoints))
	for path := range objectEndpoints {
		paths = append(paths, path)
	}

	return paths
}()

func main() {
	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range Objects() {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"urlPath", object.URLPath,
				"error", object.Problem,
			)

			continue
		}

		for _, field := range object.Fields {
			schemas.Add(common.ModuleRoot, object.ObjectName, object.DisplayName,
				"/api/v2"+object.URLPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	// The /estimates list returns reference stubs only (id, isPrimary, job)
	// and silently ignores ?includes= (verified live), so the read connector
	// hydrates every estimate from GET /estimates/{estimateId}. The detail
	// fields (financials, createdDate, sections, ...) are therefore part of
	// the estimates read shape and merged into its schema here, keyed under
	// the list's URL path and response key. Query params are intentionally
	// not registered: queryParamStats tracks list endpoints only.
	for _, field := range estimateDetailFields() {
		schemas.Add(common.ModuleRoot, "estimates", "Estimates",
			"/api/v2/estimates", "items",
			utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, nil)
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.", "objects", len(objectEndpoints))
}

// Objects extracts schemas via ReadObjects (not ReadObjectsGet) so paths with
// OpenAPI placeholders aren't dropped — the read layer needs them for fan-out.
// NestedIDPathIgnorer keeps single-level nesting and rejects 2-level paths.
func Objects() []metadatadef.Schema {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithArrayItemAutoSelection(),
		api3.WithDisplayNamePostProcessors(
			api3.SlashesToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjects("GET",
		api3.AndPathMatcher{
			api3.NewAllowPathStrategy(allowedPaths),
			api3.NestedIDPathIgnorer{},
		},
		objectEndpoints,
		displayNameOverride,
		nil,
	)
	goutils.MustBeNil(err)

	return objects
}

// estimateDetailFields returns the flattened property set of the estimate
// component schema — the single-object response of GET /estimates/{estimateId}.
// api3's ReadObjects cannot extract it: its extractors are collection-oriented
// and, pointed at the detail endpoint, auto-select the only inner array
// (sections) and return section fields instead of the estimate's own. So the
// properties are read directly from the spec.
func estimateDetailFields() []metadatadef.Field {
	doc, err := openapi3.NewLoader().LoadFromData(openapi.FileBytes())
	goutils.MustBeNil(err)

	schema, ok := doc.Components.Schemas["estimate"]
	if !ok || schema.Value == nil {
		panic("estimate component schema missing from AccuLynx OpenAPI spec")
	}

	props := flattenedProperties(schema.Value)

	names := datautils.Map[string, *openapi3.SchemaRef](props).Keys()
	sort.Strings(names)

	fields := make([]metadatadef.Field, 0, len(names))
	for _, name := range names {
		fields = append(fields, metadatadef.Field{
			Name: name,
			Type: propertyTypeName(props[name]),
		})
	}

	return fields
}

// flattenedProperties merges a schema's own properties with those of its
// allOf components, recursively. Own properties win on collisions.
func flattenedProperties(schema *openapi3.Schema) map[string]*openapi3.SchemaRef {
	out := make(map[string]*openapi3.SchemaRef)

	for _, ref := range schema.AllOf {
		if ref != nil && ref.Value != nil {
			maps.Copy(out, flattenedProperties(ref.Value))
		}
	}

	maps.Copy(out, schema.Properties)

	return out
}

// propertyTypeName resolves a property's provider type. Schemas composed via
// $ref/allOf frequently omit an explicit type; those are objects.
func propertyTypeName(ref *openapi3.SchemaRef) string {
	if ref == nil || ref.Value == nil {
		return ""
	}

	if types := ref.Value.Type; types != nil && len(types.Slice()) > 0 {
		return types.Slice()[0]
	}

	if len(ref.Value.AllOf) > 0 || len(ref.Value.Properties) > 0 {
		return "object"
	}

	return ""
}
