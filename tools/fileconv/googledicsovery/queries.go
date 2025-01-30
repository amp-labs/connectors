package googledicsovery

import (
	"net/http"
	"sort"

	"github.com/amp-labs/connectors/internal/metadatadef"
)

// Explorer allows to traverse schema in most common ways
// relevant for connectors metadata extraction.
type Explorer struct {
	document Document
	*parameters
}

func (e Explorer) ReadObjectsGet(
	displayNameOverride map[string]string,
) (metadatadef.Schemas, error) {
	return e.ReadObjects(http.MethodGet, displayNameOverride)
}

func (e Explorer) ReadObjects(
	httpMethod string,
	displayNameOverride map[string]string,
) (metadatadef.Schemas, error) {
	schemas := e.document.ListObjects(httpMethod)

	for i, schema := range schemas {
		displayName, ok := displayNameOverride[schema.ObjectName]
		if !ok {
			displayName = e.displayPostProcessing(schema.ObjectName)
		}

		schema.DisplayName = displayName
		schemas[i] = schema
	}

	sort.Slice(schemas, func(i, j int) bool {
		return schemas[i].Problem == nil && schemas[j].Problem != nil
	})

	return schemas, nil
}
