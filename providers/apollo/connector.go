package apollo

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	// Basic connector
	*components.Connector
}

type operation string

// NewConnector is an old constructor, use NewConnectorV2.
// Deprecated.
func NewConnector(opts ...Option) (*Connector, error) {
	params, err := newParams(opts)
	if err != nil {
		return nil, err
	}

	return NewConnectorV2(*params)
}

func NewConnectorV2(params common.Parameters) (*Connector, error) {
	return components.Initialize(providers.Apollo, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	return &Connector{Connector: base}, nil
}

// getAPIURL builds the url we can write/read data from
// Depending on the operation(read or write), some objects will need different endpoints.
// That's the sole purpose of the variable ops.
func (c *Connector) getAPIURL(objectName string, ops operation) (*urlbuilder.URL, error) {
	objectName = constructSupportedObjectName(objectName)

	relativePath := strings.Join([]string{restAPIPrefix, objectName}, "/")

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, relativePath)
	if err != nil {
		return nil, err
	}

	// If the given object uses search endpoint for Reading,
	// checks for the  method and makes the call.
	// currently we do not support routing to Search method.
	//
	if usesSearching(objectName) && ops == readOp {
		url.AddPath(searchingPath)
	}

	return url, nil
}
