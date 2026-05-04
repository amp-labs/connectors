package spec

import (
	"fmt"
	"strings"

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
	// URLPath: the OpenAPI path string (e.g. "/v2/affiliates/commissionPrograms").
	URLPath string
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
	// Raw exposes the underlying OpenAPI schema for advanced use-cases.
	Raw *openapi3.Schema
}

type Field struct {
	Name         string
	Type         string
	ValueOptions []string
}

type Fields = datautils.Map[string, Field]

func (s Schema) String() string {
	if s.Problem != nil {
		return fmt.Sprintf("    {%v}    ", s.URLPath)
	}

	return fmt.Sprintf("%v=[%v]", s.URLPath, strings.Join(s.Fields.Keys(), ","))
}
