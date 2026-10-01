package jira

import (
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
)

const apiVersion = "3"

type Adapter struct {
	components.BaseConnector
}

func NewAdapter(base *components.BaseConnector) (*Adapter, error) {
	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
		HTML: &interpreter.DirectFaultyResponder{Callback: interpretHTMLError},
	}.Handle

	// TODO this is temporary. Error handling should belong to operations.HTTPHandlers.ErrorHandler.
	// Therefore this adapter must use operations.HTTPHandlers when making API requests.
	// FIXME this creates the side effects and overrides the error handling from the parent.
	// This is why the error handling must be local and not connected to the Auth clients.
	base.HTTPClient().ErrorHandler = errorHandler

	return &Adapter{
		BaseConnector: *base,
	}, nil
}

// URL format for providers.ModuleAtlassianJira follows structure applicable to Oauth2 Atlassian apps:
// https://developer.atlassian.com/cloud/jira/platform/rest/v2/intro/#other-integrations
func (a *Adapter) getModuleURL(path string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, path)
}
