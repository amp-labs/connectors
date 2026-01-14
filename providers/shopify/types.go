package shopify

// MetadataResponse represents the GraphQL introspection response structure.
// nolint
type MetadataResponse struct {
	Data struct {
		Type struct {
			Name   string  `json:"name"`
			Fields []Field `json:"fields"`
		} `json:"__type"`
	} `json:"data"`
}

// Field represents a field in the GraphQL type.
// nolint
type Field struct {
	Name string `json:"name"`
	Type struct {
		Name   string `json:"name"`
		Kind   string `json:"kind"`
		OfType *struct {
			Name string `json:"name"`
			Kind string `json:"kind"`
		} `json:"ofType"`
	} `json:"type"`
}

// WriteResponse represents the GraphQL mutation response structure.
// // nolint
type WriteResponse struct {
	Data map[string]map[string]any `json:"data"`
}

// UserError represents an error returned by a Shopify mutation.
// nolint
type UserError struct {
	Field   []string `json:"field"`
	Message string   `json:"message"`
}

// MutationData is a helper struct for parsing mutation responses.
// nolint
type MutationData struct {
	UserErrors []UserError
	Object     map[string]any
}
