package core

// PatchPayload represents a write payload that contains a JSON Patch array.
//
// The `patch` field holds a list of JSON Patch operations (add/replace/remove)
// targeting specific paths in the ConnectWise object. When this structure is
// detected, the connector uses HTTP PATCH and sends the inner array as the
// request body.
type PatchPayload struct {
	Patch []PatchOperationPayload `json:"patch"`
}

// PatchOperationPayload represents a single JSON Patch operation as defined by RFC 6902.
type PatchOperationPayload struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value,omitempty"`
}
