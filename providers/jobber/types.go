// nolint
package jobber

type TypeKind string

const (
	KindScalar  TypeKind = "SCALAR"
	KindObject  TypeKind = "OBJECT"
	KindNonNull TypeKind = "NON_NULL"
	KindList    TypeKind = "LIST"
)

// TypeInfo represents the type information in the GraphQL schema.
type TypeInfo struct {
	Name   string     `json:"name"`
	Kind   TypeKind   `json:"kind"`
	OfType OfTypeInfo `json:"ofType"`
}

type OfTypeInfo struct {
	Name string   `json:"name"`
	Kind TypeKind `json:"kind"`
}

// Field represents a field in the GraphQL schema.
type Field struct {
	Name string   `json:"name"`
	Type TypeInfo `json:"type"`
}

// TypeMetadata represents the type metadata in the GraphQL schema.
type TypeMetadata struct {
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

// MetadataResponse represents the response structure for metadata queries.
// nolint
type MetadataResponse struct {
	Data struct {
		Type TypeMetadata `json:"__type"`
	} `json:"data"`
}

type Response struct {
	Errors any          `json:"errors"`
	Data   ResponseData `json:"data"`
}

type ResponseData struct {
	Users       []map[string]any `json:"users,omitempty"`
	Transcripts []map[string]any `json:"transcripts,omitempty"`
	Bites       []map[string]any `json:"bites,omitempty"`
}
