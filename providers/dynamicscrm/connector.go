package dynamicscrm

import (
	"fmt"

	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "v9.2"

type Connector struct {
	deep.Clients
	deep.EmptyCloser
}

type parameters struct {
	paramsbuilder.Client
	paramsbuilder.Workspace
}

func NewConnector(opts ...Option) (*Connector, error) {
	return deep.Connector[Connector, parameters](providers.DynamicsCRM, interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}).Build(opts)
}

func (c *Connector) getURL(arg string) (*urlbuilder.URL, error) {
	return constructURL(c.BaseURL(), apiVersion, arg)
}

func (c *Connector) getEntityDefinitionURL(arg naming.SingularString) (*urlbuilder.URL, error) {
	// This endpoint returns schema of an object.
	// Schema name must be singular.
	path := fmt.Sprintf("EntityDefinitions(LogicalName='%v')", arg.String())

	return c.getURL(path)
}

func (c *Connector) getEntityAttributesURL(arg naming.SingularString) (*urlbuilder.URL, error) {
	// This endpoint will describe attributes present on schema and its properties.
	// Schema name must be singular.
	path := fmt.Sprintf("EntityDefinitions(LogicalName='%v')/Attributes", arg.String())

	return c.getURL(path)
}
