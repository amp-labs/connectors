package aws

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient
	common.RequireModule
	common.RequireMetadata

	// supported operations
	components.Reader

	region          string
	identityStoreId string
	instanceARN     string
}

func NewConnector(params common.Parameters) (*Connector, error) {
	conn, err := components.Initialize(providers.AWS, params,
		func(connector *components.Connector) (*Connector, error) {
			var expectedMetadataKeys []string
			if params.Module == providers.ModuleAWSIdentityCenter {
				expectedMetadataKeys = []string{"region", "identityStoreId", "instanceARN"}
			}

			return constructor(connector, expectedMetadataKeys)
		},
	)
	if err != nil {
		return nil, err
	}

	// TODO this should be resolved by the ProviderInfo initialization.
	conn.region = params.Metadata["region"]
	conn.identityStoreId = params.Metadata["identityStoreId"]
	conn.instanceARN = params.Metadata["instanceARN"]

	return conn, nil
}

func constructor(base *components.Connector, expectedMetadataKeys []string) (*Connector, error) {
	connector := &Connector{
		Connector: base,
		RequireModule: common.RequireModule{
			ExpectedModules: []common.ModuleID{
				providers.ModuleAWSIdentityCenter,
			},
		},
		RequireMetadata: common.RequireMetadata{
			ExpectedMetadataKeys: expectedMetadataKeys,
		},
	}

	return connector, nil
}

// nolint:unused
func (c *Connector) getModuleURL() (string, error) {
	modules := c.ProviderInfo().Modules
	if modules == nil {
		return "", common.ErrInvalidModuleDeclaration
	}

	baseURL := (*modules)[c.Module()].BaseURL
	baseURL = strings.Replace(baseURL, "{{.region}}", c.region, 1)

	return baseURL, nil
}
