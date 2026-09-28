package reducers

import (
	"errors"
	"strings"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
)

// ShortestNameFromUrl -- algorithm (progressive disambiguation).
func ShortestNameFromUrl(schemas []spec.Schema) []spec.Schema {
	type State struct {
		schema *spec.Schema // need side effects, we modify object name
		parts  []string
		index  int // current position from the right
	}

	states := make([]State, len(schemas))

	for index, schema := range schemas {
		parts := strings.Split(strings.Trim(schema.UrlPath, "/"), "/")
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

// ShortestNameFromUrlWithFullFallback assigns each schema an ObjectName derived from
// the last segment of its UrlPath. If multiple schemas share the same name, all
// colliding schemas use their full URL path (with leading/trailing slashes trimmed)
// as the ObjectName.
func ShortestNameFromUrlWithFullFallback(schemas []spec.Schema) []spec.Schema {
	// Work on copies so we can mutate ObjectName
	result := make([]spec.Schema, len(schemas))
	copy(result, schemas)

	// First pass: assign last segment as ObjectName
	for i := range result {
		parts := strings.Split(strings.Trim(result[i].UrlPath, "/"), "/")
		if len(parts) > 0 {
			result[i].ObjectName = parts[len(parts)-1]
		} else {
			result[i].Problem = errors.New("url is empty after splitting using '/'")
		}
	}

	// Group by ObjectName to detect collisions
	groups := map[string][]int{} // name -> indices in result
	for idx, schema := range result {
		groups[schema.ObjectName] = append(groups[schema.ObjectName], idx)
	}

	// Second pass: for any name that appears more than once, use full URL path
	for _, indices := range groups {
		if len(indices) <= 1 {
			continue
		}
		for _, idx := range indices {
			// Full URL path, trimmed of leading/trailing slashes
			url := result[idx].UrlPath
			result[idx].ObjectName = strings.Trim(url, "/")
		}
	}

	return result
}
