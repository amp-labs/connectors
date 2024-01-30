package connectors

import (
	"fmt"
	"io"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

// BasicConnector is an interface that can be used to implement a connector with
// basic configuration about the provider.
type BasicConnector interface {
	fmt.Stringer
	io.Closer

	// Name returns the name of the connector.
	Name() string

	// HTTPClient returns the underlying HTTP client. This is useful for proxy requests.
	HTTPClient() *common.HTTPClient

	// Provider returns the connector provider.
	Provider() providers.Provider
}
