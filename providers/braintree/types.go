package braintree

// GraphQLError represents an error in the GraphQL response.
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLErrorLocation `json:"locations,omitempty"`
	Path       []string               `json:"path,omitempty"`
	Extensions map[string]any         `json:"extensions,omitempty"`
}

type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// ResponseError represents the error structure in GraphQL responses.
type ResponseError struct {
	Errors []GraphQLError `json:"errors,omitempty"`
}
