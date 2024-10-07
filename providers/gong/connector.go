package gong

import (
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/gong/metadata"
)

const ApiVersion = "v2"

type Connector struct {
	deep.Clients
	deep.EmptyCloser
	deep.StaticMetadata
}

type parameters struct {
	paramsbuilder.Client
}

func NewConnector(opts ...Option) (*Connector, error) {
	return deep.Connector[Connector, parameters](providers.Gong, interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}).Setup(func(conn *Connector) {
		conn.StaticMetadata = deep.NewStaticMetadata(metadata.Schemas)
	}).Build(opts)
}

func (c *Connector) getURL(arg string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL(), ApiVersion, arg)
}
