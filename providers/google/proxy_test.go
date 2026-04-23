package google

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestProxy(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	tests := []testroutines.Proxy{
		{
			Name: "Google Calendar Proxy",
			Builder: func() (connectors.ProxyConnector, error) {
				return constructTestCalendarConnector("")
			},
			ExpectedProxy: &connectors.ProxyConfig{
				URL: "https://www.googleapis.com",
			},
			ExpectedModuleProxy: &connectors.ProxyConfig{
				URL: "https://www.googleapis.com/calendar",
			},
		},
		{
			Name: "Google Contacts Proxy",
			Builder: func() (connectors.ProxyConnector, error) {
				return constructTestContactsConnector("")
			},
			ExpectedProxy: &connectors.ProxyConfig{
				URL: "https://www.googleapis.com",
			},
			ExpectedModuleProxy: &connectors.ProxyConfig{
				URL: "https://people.googleapis.com",
			},
		},
		{
			Name: "Google Mail Proxy",
			Builder: func() (connectors.ProxyConnector, error) {
				return constructTestMailConnector("")
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
