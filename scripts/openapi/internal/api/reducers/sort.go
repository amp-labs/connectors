package reducers

import (
	"sort"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
)

func SortByObjectName(schemas []spec.Schema) []spec.Schema {
	result := make([]spec.Schema, len(schemas))

	// Copy to avoid mutating the original slice (pure reducer semantics).
	copy(result, schemas)

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].ObjectName < result[j].ObjectName
	})

	return result
}
