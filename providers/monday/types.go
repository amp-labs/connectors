package monday

type TypeKind string

const (
	KindScalar  TypeKind = "SCALAR"
	KindObject  TypeKind = "OBJECT"
	KindEnum    TypeKind = "ENUM"
	KindNonNull TypeKind = "NON_NULL"
	KindList    TypeKind = "LIST"
)

// TypeInfo represents the type information in the GraphQL schema.
type TypeInfo struct {
	Name   *string  `json:"name"`
	Kind   TypeKind `json:"kind"`
	OfType *struct {
		Name string `json:"name"`
	} `json:"ofType"`
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

// MondayResponse represents the top-level response structure from Monday.com API.
type MondayResponse struct {
	Data ResponseData `json:"data"`
}

// ResponseData contains the different object types that can be returned.
type ResponseData struct {
	Users  []User  `json:"users,omitempty"`
	Boards []Board `json:"boards,omitempty"`
}

// User represents a user in the Monday.com API.
type User struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// Board represents a board in the Monday.com API.
type Board struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	State       string `json:"state"`
	Permissions string `json:"permissions"`
	ItemsCount  int    `json:"itemsCount"`
	Type        string `json:"type"`
	UpdatedAt   string `json:"updatedAt"`
	URL         string `json:"url"`
	WorkspaceID string `json:"workspaceId"`
}

// MondayRecord represents a record in the Monday.com API response
// The fields are kept as map[string]any since they vary by object type.
type MondayRecord struct {
	ID   string         `json:"id,omitempty"`
	Name string         `json:"name,omitempty"`
	Data map[string]any `json:"-"` // Catch-all for other fields
}
