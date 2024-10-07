package intercom

import (
	"github.com/amp-labs/connectors/internal/deep"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "2.11"

var apiVersionHeader = common.Header{ // nolint:gochecknoglobals
	Key:   "Intercom-Version",
	Value: apiVersion,
}

type Connector struct {
	deep.Clients
	deep.EmptyCloser
}

type parameters struct {
	paramsbuilder.Client
}

func NewConnector(opts ...Option) (*Connector, error) {
	return deep.Connector[Connector, parameters](providers.Intercom, interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}).Build(opts)
}

func (c *Connector) getURL(arg string) (*urlbuilder.URL, error) {
	return constructURL(c.BaseURL(), arg)
}
