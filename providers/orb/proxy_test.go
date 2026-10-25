package orb

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestProxy(t *testing.T) {
	t.Parallel()

	tests := []testconn.Proxy{
		{
			Name: "Proxy targets the catalog BaseURL without a version",
			Builder: func() (connectors.ProxyConnector, error) {
				return NewConnector(common.ConnectorParams{
					Module:              common.ModuleRoot,
					AuthenticatedClient: mockutils.NewClient(),
				})
			},
			ExpectedProxy: &connectors.ProxyConfig{
				URL: "https://api.withorb.com",
			},
			ExpectedModuleProxy: &connectors.ProxyConfig{
				URL: "https://api.withorb.com",
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
