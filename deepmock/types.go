package deepmock

// SchemaConfig holds the object name and raw JSON schema bytes.
type SchemaConfig struct {
	ObjectName string
	RawSchema  []byte
}

// RecordID is an alias for string representing a record identifier.
type RecordID string

// ObjectName is an alias for string representing an object name.
type ObjectName string
