package connectors

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/salesforce"
)

// Connector is an interface that all connectors must implement.
type Connector interface {
	Read(ctx context.Context, params ReadParams) (*ReadResult, error)
}

// API is a function that returns a Connector. It's used as a factory.
type API[Token any, Conn Connector] func(workspaceRef string, getToken func() (Token, error)) Conn

var (
	// Salesforce is an API that returns a new Salesforce Connector.
	Salesforce API[string, *salesforce.Connector] = salesforce.NewConnector
)

// We re-export the following types so that they can be used by consumers of this library.
type ReadParams = common.ReadParams
type ReadResult = common.ReadResult
type ErrorWithStatus = common.ErrorWithStatus

// New returns a new Connector.
func New[T any, Conn Connector](api API[T, Conn], workspaceRef string, getToken func() (T, error)) Connector {
	return api(workspaceRef, getToken)
}
