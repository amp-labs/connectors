// Package objects answers "does provider X support operation O on object Y".
//
// Nothing answered that before. Support was declared in three places that do not agree and
// none of which is queryable: 138 connectors declare operation support in Go, 78 ship a
// schemas.json, a couple of dozen READMEs carry hand-maintained tables, and the generated
// catalog stops at provider and module granularity. So the only way to find out whether a
// provider could write Task was to try it.
package objects

// Operation is an action that can be performed on an object.
type Operation string

// The operations the matrix records. These mirror the provider-level Support flags.
const (
	OperationRead      Operation = "read"
	OperationWrite     Operation = "write"
	OperationDelete    Operation = "delete"
	OperationSubscribe Operation = "subscribe"
)

// Source records where an object's entry came from, so a caller can tell a fact from a
// claim.
type Source string

const (
	// SourceSchema means the object is in the provider's shipped schemas.json, which is
	// generated from the provider's own API.
	SourceSchema Source = "schema"

	// SourceDeclared means a human wrote it down. Used for object-generic providers,
	// where there is no fixed list to generate from and the enumerated objects are a
	// convenience rather than a boundary.
	SourceDeclared Source = "declared"
)

// ObjectSupport is what can be done with one object.
type ObjectSupport struct {
	Read      bool   `json:"read"`
	Write     bool   `json:"write"`
	Delete    bool   `json:"delete"`
	Subscribe bool   `json:"subscribe"`
	Source    Source `json:"source"`
}

// Supports reports whether the given operation is supported.
func (o ObjectSupport) Supports(operation Operation) bool {
	switch operation {
	case OperationRead:
		return o.Read
	case OperationWrite:
		return o.Write
	case OperationDelete:
		return o.Delete
	case OperationSubscribe:
		return o.Subscribe
	default:
		return false
	}
}

// ProviderObjects is one provider's entry in the matrix.
type ProviderObjects struct {
	Provider string `json:"provider"`

	// Generic is true when the provider works against whatever object the credential can
	// reach rather than a fixed list -- Salesforce, HubSpot, Dynamics and Close all behave
	// this way. For these, Objects is a list of objects known to work, not a limit: an
	// object absent from it may still be reachable.
	//
	// The distinction matters because "not listed" means two opposite things depending on
	// this flag: unsupported for a fixed-list provider, unknown for a generic one.
	Generic bool `json:"generic"`

	// Objects is keyed by module, then by object name.
	Objects map[string]map[string]ObjectSupport `json:"objects"`
}

// Matrix is the generated provider-by-object support matrix.
type Matrix struct {
	// Timestamp is when the matrix was generated.
	Timestamp string `json:"timestamp"`

	// Providers is keyed by provider name.
	Providers map[string]ProviderObjects `json:"providers"`
}
