package components

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

// Transport
// TODO: Add support for XML, CSV, etc.
type Transport struct {
	ProviderContext

	json *common.JSONHTTPClient
}

// NewTransport
// TODO: The JSON client by itself is not providing any functionality right now - this is to only provide
// continuity for the existing codebase. We should refactor the existing JSON/XML/CSV/HTTP clients to
// satisfy a common interface, and then hook them up in here.
func NewTransport(
	provider providers.Provider,
	params common.ConnectorParams,
) (*Transport, error) {
	providerContext, err := NewProviderContext(provider, params.Module, params.Workspace, params.Metadata)
	if err != nil {
		return nil, err
	}

	return &Transport{
		ProviderContext: *providerContext,
		json: &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Base:   providerContext.ProviderInfo().BaseURL,
				Client: params.AuthenticatedClient,

				// ErrorHandler is set to a default, but can be overridden using options.
				ErrorHandler: common.InterpretError,

				// No ResponseHandler is set, but can be overridden using options.
			},
			ErrorPostProcessor: common.ErrorPostProcessor{},
		},
	}, nil
}

// SetBaseURL should be used for setting up unit tests.
// To better indicate the intent use SetUnitTestMockServerBaseURL.
// Deprecated.
func (t *Transport) SetBaseURL(newURL string) {
	t.ProviderContext.providerInfo.BaseURL = newURL
	t.ProviderContext.moduleInfo.BaseURL = newURL
	t.json.HTTPClient.Base = newURL
}

// SetUnitTestBaseURL should be used for setting up unit tests.
// To better handle diverse Module vs Provider BaseURLs use SetUnitTestMockServerBaseURL.
// Deprecated.
func (t *Transport) SetUnitTestBaseURL(newURL string) {
	t.ProviderContext.providerInfo.BaseURL = newURL
	t.ProviderContext.moduleInfo.BaseURL = newURL
	t.json.HTTPClient.Base = newURL
}

// SetUnitTestMockServerBaseURL replaces the URL Origin with mock server URL Origin.
// This allows to reroute all requests to mock server used in unit tests and preserve all URI parts if any.
func (t *Transport) SetUnitTestMockServerBaseURL(testServerURL string) {
	providerURL := t.ProviderContext.providerInfo.BaseURL
	t.ProviderContext.providerInfo.BaseURL = mockutils.ReplaceURLOrigin(providerURL, testServerURL)
	moduleURL := t.ProviderContext.moduleInfo.BaseURL
	t.ProviderContext.moduleInfo.BaseURL = mockutils.ReplaceURLOrigin(moduleURL, testServerURL)
	t.json.HTTPClient.Base = testServerURL
}

func (t *Transport) SetErrorHandler(handler common.ErrorHandler) {
	t.HTTPClient().ErrorHandler = handler
}

func (t *Transport) JSONHTTPClient() *common.JSONHTTPClient { return t.json }
func (t *Transport) HTTPClient() *common.HTTPClient         { return t.json.HTTPClient }
