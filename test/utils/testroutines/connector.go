// Package testroutines holds a collection of common test procedures.
// They provide a framework to write mock tests.
package testroutines

import (
	"regexp"
	"testing"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

// ConnectorBuilder is a callback method to construct and configure connector for testing.
// This is a factory method called for every test suite.
type ConnectorBuilder[Conn any] func() (Conn, error)

func (builder ConnectorBuilder[C]) Build(t *testing.T, testCaseName string) C {
	conn, err := builder()
	if err != nil {
		t.Fatalf("%s: error in test while constructing connector %v", testCaseName, err)
	}

	return conn
}

// OverrideURLOrigin changes the URL origin.
// For testing purposes API calls will be redirected to the mock server.
func OverrideURLOrigin(manager *components.URLManager, providerInfo *providers.ProviderInfo, originURL string) {
	// TODO no connector should rely on providers.ProviderInfo directly.
	providerInfo.BaseURL = originURL + extractPath(manager.RootAPI.Format)

	manager.RootAPI.Format = originURL + extractPath(manager.RootAPI.Format)
	manager.ModuleAPI.Format = originURL + extractPath(manager.ModuleAPI.Format)
}

func extractPath(urlTemplate string) string {
	var pathRegex = regexp.MustCompile(`https?://[^/]+(/.*)`)

	matches := pathRegex.FindStringSubmatch(urlTemplate)
	if len(matches) < 2 {
		// no path found in template
		return ""
	}

	return matches[1]
}
