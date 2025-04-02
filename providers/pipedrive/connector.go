package pipedrive

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/pipedrive/metadata"
)

const (
	apiVersion string = "v1"    // nolint:gochecknoglobals
	limitQuery string = "limit" // nolint:gochecknoglobals
)

// Connector represents the Pipedrive Connector.
type Connector struct {
	// Basic connector
	*components.Connector
}

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
	return components.Initialize(providers.Pipedrive, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	return &Connector{Connector: base}, nil
}

// getAPIURL constructs a specific object's resource URL in the format
// `{{baseURL}}/{{version}}/{{objectName}}`.
func (c *Connector) getAPIURL(arg string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, arg)
}

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupURLPath(c.Module(), objectName)
	if err != nil {
		return nil, err
	}

	return c.getAPIURL(path)
}
