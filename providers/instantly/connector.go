package instantly

import (
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers/instantly/metadata"

	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "v1"

type Connector struct {
	deep.Clients
	deep.EmptyCloser
	deep.StaticMetadata
}

type parameters struct {
	paramsbuilder.Client
	// Error is set when any With<Method> fails, used for parameters validation.
	setupError error
}

func NewConnector(opts ...Option) (*Connector, error) {
	return deep.Connector[Connector, parameters](providers.Instantly, interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}).Setup(func(conn *Connector) {
		conn.StaticMetadata = deep.NewStaticMetadata(metadata.Schemas)
	}).Build(opts)
}

func (c *Connector) getURL(parts ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL(), append([]string{
		apiVersion,
	}, parts...)...)
}
