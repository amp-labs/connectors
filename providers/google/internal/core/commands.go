package core

import (
	"strings"

	"github.com/amp-labs/connectors/common"
)

type Endpoints map[OperationName]map[string]Endpoint

type OperationName string

const (
	OperationCreate OperationName = "CREATE"
	OperationUpdate OperationName = "UPDATE"
	OperationDelete OperationName = "DELETE"
)

// Find looks up the endpoint for the object.
// The resulting endpoint will have fully formed URL respecting object name and record id which could be empty.
func (o Endpoints) Find(operationName OperationName, objectName, recordID string) (*Endpoint, error) {
	registry := o[operationName]

	operation, ok := registry[objectName]
	if !ok {
		return nil, common.ErrObjectNotSupported
	}

	return &Endpoint{
		Method: operation.Method,
		Path:   operation.resolvePath(recordID),
	}, nil
}

type Endpoint struct {
	Method string
	Path   string
}

func (d Endpoint) resolvePath(recordID string) string {
	if len(recordID) == 0 {
		// Usually this is a create or command endpoint.
		return d.Path
	}

	// No template. Usually record identifier is attached at the end of endpoint.
	if !strings.Contains(d.Path, "{{.recordID}}") {
		return d.Path + "/" + recordID
	}

	// Insert recordID inside URL according to the template format.
	return strings.ReplaceAll(d.Path, "{{.recordID}}", recordID)
}
