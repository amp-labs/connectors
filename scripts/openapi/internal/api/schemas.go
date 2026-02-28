package api

import (
	"net/http"
	"sort"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api/document"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
	"github.com/getkin/kin-openapi/openapi3"
)

type HTTPOperation = document.HTTPOperation

const (
	GET    HTTPOperation = http.MethodGet
	POST   HTTPOperation = http.MethodPost
	PUT    HTTPOperation = http.MethodPut
	DELETE HTTPOperation = http.MethodDelete
)

type Extractor struct {
	doc *document.Document
}

func NewExtractor(data *openapi3.T) *Extractor {
	return &Extractor{
		doc: document.New(data),
	}
}

func (e Extractor) ExtractListSchemas(httpOperation HTTPOperation, opts ...ListOption) ([]spec.Schema, error) {
	params := createParams(opts)
	schemas := make([]spec.Schema, 0)

	for _, path := range e.doc.GetPaths() {
		schema, found, err := path.RetrieveSchemaOperation(httpOperation,
			params.locator,
			params.propertyFlattener,
			params.mediaType,
			*params.autoSelectArrayItem,
		)
		if err != nil {
			return nil, err
		}

		if found {
			// schema was found save it
			schemas = append(schemas, *schema)
		}
	}

	sort.Slice(schemas, func(i, j int) bool {
		return schemas[i].Problem == nil && schemas[j].Problem != nil
	})

	return schemas, nil
}
