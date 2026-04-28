package salesforce

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestProxy(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	tests := []testroutines.Proxy{
		{
			Name: "Salesforce CRM Proxy",
			Builder: func() (connectors.ProxyConnector, error) {
				return constructTestConnector("")
			},
			ExpectedProxy: &connectors.ProxyConfig{
				URL: "https://test-workspace.my.salesforce.com",
			},
			ExpectedModuleProxy: &connectors.ProxyConfig{
				URL: "https://test-workspace.my.salesforce.com",
			},
		},
		{
			Name: "Salesforce Account Engagement Proxy",
			Builder: func() (connectors.ProxyConnector, error) {
				return constructTestConnectorAccountEngagement("")
			},
			ExpectedProxy: &connectors.ProxyConfig{
				URL: "https://test-workspace.my.salesforce.com",
			},
			ExpectedModuleProxy: &connectors.ProxyConfig{
				URL: "https://pi.pardot.com",
			},
		},
		{
			Name: "Salesforce Account Engagement Demo Proxy",
			Builder: func() (connectors.ProxyConnector, error) {
				return constructTestConnectorAccountEngagementDemo("")
			},
			ExpectedProxy: &connectors.ProxyConfig{
				URL: "https://test-workspace.my.salesforce.com",
			},
			ExpectedModuleProxy: &connectors.ProxyConfig{
				URL: "https://pi.demo.pardot.com",
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
