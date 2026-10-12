package mail

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestProxy(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	tests := []testconn.Proxy{
		{
			Name: "Google Mail Proxy",
			Builder: func() (connectors.ProxyConnector, error) {
				return constructTestAdapter(mockserver.Dummy())
			},
			ExpectedProxy: &connectors.ProxyConfig{
				URL: "https://www.googleapis.com",
			},
			ExpectedModuleProxy: &connectors.ProxyConfig{
				URL: "https://gmail.googleapis.com/gmail",
			},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t)
		})
	}
}
