package netsuite

// metadataResponse is the response from the /metadata-catalog/{object} endpoint.
type metadataResponse struct {
	// The type of the object.
	Type string `json:"type"`

	// Contains key-value pairs of field names and their metadata.
	// Metadata can have title, type,description and nullable, even properties again.
	Properties map[string]fieldMetadata `json:"properties"`

	// The fields that the object can be filtered by.
	Filterable []string `json:"x-ns-filterable"`
}

type fieldMetadata struct {
	Title  string `json:"title"`
	Type   string `json:"type"`   // string, number, boolean, object, array, null
	Format string `json:"format"` // date-time, date, etc.
}
