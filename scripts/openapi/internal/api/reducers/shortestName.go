package reducers

import (
	"strings"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
)

// ShortestNameFromURL -- algorithm (progressive disambiguation).
func ShortestNameFromURL(schemas []spec.Schema) []spec.Schema {
	type State struct {
		schema *spec.Schema // need side effects, we modify object name
		parts  []string
		index  int // current position from the right
	}

	states := make([]State, len(schemas))

	for index, schema := range schemas {
		parts := strings.Split(strings.Trim(schema.URLPath, "/"), "/")
		states[index] = State{
			schema: &schema,
			parts:  parts,
			index:  len(parts) - 1, // start from last segment
		}

		states[index].schema.ObjectName = parts[states[index].index]
	}

	unique := false
	for !unique {
		// Group by current ObjectName
		groups := map[string][]*State{}
		for index := range states {
			groups[states[index].schema.ObjectName] = append(groups[states[index].schema.ObjectName], &states[index])
		}

		unique = true

		for _, group := range groups {
			if len(group) > 1 {
				unique = false

				for _, state := range group {
					// If there are more URI parts increase the ObjectName by adding it as prefix.
					// When index is 0 it means that object name matches full URL (first slash trimmed).
					if state.index > 0 {
						state.index--
						state.schema.ObjectName = state.parts[state.index] + "/" + state.schema.ObjectName
					}
				}
			}
		}
	}

	// Extract final schemas
	result := make([]spec.Schema, len(states))
	for index, state := range states {
		result[index] = *state.schema
	}

	return result
}
