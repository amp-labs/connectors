package reply

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

// apiVersion is the URL prefix shared by every Reply V3 endpoint.
// https://docs.reply.io/api-reference
const apiVersion = "v3"

type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Reply, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	return &Connector{
		Connector: base,
	}, nil
}
