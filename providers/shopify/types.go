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
