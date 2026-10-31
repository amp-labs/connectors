package twenty_test

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/connector"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

// The connector is built the way the server builds it, through connector.New, so the test also covers its
// registration.
func TestProxy(t *testing.T) {
	t.Parallel()

	tests := []testconn.Proxy{
		{
			Name: "Proxy targets the instance root, so /rest, /graphql and /rest/metadata are all reachable",
			Builder: func() (connectors.ProxyConnector, error) {
				conn, err := connector.New(providers.Twenty, common.ConnectorParams{
					Module:              common.ModuleRoot,
					AuthenticatedClient: mockutils.NewClient(),
					Workspace:           "crm.example.com",
				})
				if err != nil {
					return nil, err
				}

				return conn.(connectors.ProxyConnector), nil //nolint:forcetypeassert
			},
			ExpectedProxy: &connectors.ProxyConfig{
				URL: "https://crm.example.com",
			},
			ExpectedModuleProxy: &connectors.ProxyConfig{
				URL: "https://crm.example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t)
		})
	}
}

func TestInstanceDomainIsRequired(t *testing.T) {
	t.Parallel()

	_, err := connector.New(providers.Twenty, common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err == nil {
		t.Fatal("expected an error without the instance domain, got none")
	}
}
