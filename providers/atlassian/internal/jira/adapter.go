package jira

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "3"

type Adapter struct {
	*components.Connector
}

func NewAdapter(params common.ConnectorParams) (*Adapter, error) {
	return components.Init(providers.Atlassian, params, constructor)
}

func constructor(_ common.ConnectorParams, base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{
		Connector: base,
	}

	adapter.SetErrorHandler(interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
		HTML: &interpreter.DirectFaultyResponder{Callback: interpretHTMLError},
	}.Handle)

	return adapter, nil
}

// URL format for providers.ModuleAtlassianJira follows structure applicable to Oauth2 Atlassian apps:
// https://developer.atlassian.com/cloud/jira/platform/rest/v2/intro/#other-integrations
func (a *Adapter) getModuleURL(path ...string) (*urlbuilder.URL, error) {
	path = append([]string{apiVersion}, path...)

	return urlbuilder.New(a.ModuleInfo().BaseURL, path...)
}
