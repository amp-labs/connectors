package connectors

import (
	"fmt"
	"io"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/hubspot"
	"github.com/amp-labs/connectors/linkedin"
	"github.com/amp-labs/connectors/salesforce"
)

// ErrInvalidAPIName is returned when the api name is invalid.
var ErrInvalidAPIName = fmt.Errorf("invalid api name")

// SimpleConnector is an interface that can be used to implement a connector with
// basic configuration about the provider.
type SimpleConnector interface {
	fmt.Stringer
	io.Closer

	// Name returns the name of the connector.
	Name() string

	// HTTPClient returns the underlying HTTP client. This is useful for proxy requests.
	HTTPClient() *common.HTTPClient
}

// NewSimple returns a new SimpleConnector.
func NewSimple(apiName string, opts map[string]any) (SimpleConnector, error) { //nolint:ireturn
	// the salesforce & hubspot connectors implement the SimpleConnector interface
	if strings.EqualFold(apiName, salesforce.Name) {
		return newSalesforce(opts)
	}

	if strings.EqualFold(apiName, hubspot.Name) {
		return newHubspot(opts)
	}

	if strings.EqualFold(apiName, "linkedin") {
		return newLinkedIn()
	}

	return nil, fmt.Errorf("%w: %s", ErrInvalidAPIName, apiName)
}

// newLinkedIn returns a new LinkedIn SimpleConnector.
func newLinkedIn() (SimpleConnector, error) { //nolint:ireturn
	return linkedin.NewSimpleConnector()
}
