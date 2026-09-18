package spec

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/getkin/kin-openapi/openapi3"
)

// Schema represents a single logical object extracted from an OpenAPI document.
// It is the core unit passed through pipelines (Filter/Map/Reduce) and is designed
// to expose *all* relevant information needed to analyze, transform, or generate
// data from an API specification.
//
// A Schema is created from an OpenAPI Operation + its response schema.
// One Schema corresponds to one endpoint.
//
// Pipelines are free to modify ObjectName, DisplayName, Fields, or any other field.
type Schema struct {
	// ObjectName: initialized to the endpoint's full URL path.
	// Pipelines commonly normalize or shorten this.
	ObjectName string
	// UrlPath: the OpenAPI path string (e.g. "/v2/affiliates/commissionPrograms").
	UrlPath string
	// Operation the HTTP verb associated with this schema (GET, POST, etc.).
	Operation string
	// DisplayName taken from the OpenAPI schema's `title` field, if present.
	DisplayName string
	// ResponseKey name of the JSON property that contains list items,
	// determined by the configured document.ArrayLocator.
	ResponseKey string
	// Fields are recursively expanded from the OpenAPI schema found at ResponseKey.
	Fields Fields
	// QueryParams lists all query parameters declared for this endpoint.
	QueryParams []string
	// Problem holds a non-fatal extraction error (e.g. couldn't find "list schema", unknown schema type).
	Problem error
	// Origin exposes the underlying OpenAPI schema for advanced use-cases.
	Origin SchemaOrigin
}

type SchemaOrigin struct {
	ResponseSchema    *openapi3.Schema
	ItemsSchemaRef    string
	ItemsSchema       *openapi3.Schema
	ItemComponentName string
}

func (s Schema) String() string {
	if s.Problem != nil {
		return fmt.Sprintf("    {%v}    ", s.UrlPath)
	}

	return fmt.Sprintf("%v=[%v]", s.UrlPath, strings.Join(s.Fields.Keys(), ","))
}

type Field struct {
	// URLPath is the endpoint path of the object that contains this field.
	URLPath string
	// Name is the field name.
	Name string
	// DisplayName is the human-readable label for the field.
	DisplayName string
	// Type is the field type as defined in the OpenAPI document.
	Type string
	// ValueType is the connector-specific logical type for this field.
	// If set, ValueType overrides any automatic inference from Type.
	// Leave empty to use the default mapping from Type to connector terminology.
	ValueType common.ValueType
	// ValueOptions contains the enumeration values defined for the field in the OpenAPI document.
	ValueOptions []string
	// ReferenceTo list of object names that this field relates to.
	ReferenceTo []string
	// Origin exposes the underlying OpenAPI schema for advanced use-cases.
	Origin FieldOrigin
}

type Fields = datautils.Map[string, Field]

type FieldOrigin struct {
	// ItemsSchema is the top-level schema this field was expanded from.
	ItemsSchema *openapi3.Schema

	// Schema is the immediate OpenAPI schema object where this field was defined.
	// For simple properties this is often the same as ItemsSchema; for referenced
	// components it may be a different schema (e.g. CampaignStatusReference).
	// See RefPath for a complete OpenAPI path from items schema to schema owning this field.
	Schema *openapi3.Schema

	// RefPath is the chain of $ref strings traversed to reach this field.
	// Example:
	//   ["#/components/schemas/Campaign", "#/components/schemas/CampaignStatusReference"]
	// May be empty if everything is inlined within Schema.
	RefPath []string

	// ComponentRefs is the list of component $refs that define this field's type.
	// Each element is a full $ref, e.g. "#/components/schemas/GroupReference".
	// Length:
	//   0 → fully inlined (no named components)
	//   1 → single-component reference (common case)
	//   >1 → polymorphic (oneOf/anyOf/allOf with multiple distinct components)
	ComponentRefs []string

	// ComponentNames is the list of OpenAPI component names that define this field's type.
	// Each element is the last segment of the corresponding $ref in ComponentRefs,
	// e.g. "GroupReference" for "#/components/schemas/GroupReference".
	// ComponentNames has the same length and order as ComponentRefs.
	ComponentNames []string
}
