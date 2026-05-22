// Package gotoconn implements the GoTo connector.
//
// The package is named "gotoconn" instead of "goto" because "goto" is a
// reserved keyword in Go and cannot be used as a package identifier. The
// "conn" suffix is short for "connector".
package gotoconn

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client & account
	common.RequireAuthenticatedClient
	common.PostAuthInfo

	accountKey string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	conn, err := components.Initialize(providers.GoTo, params, constructor)
	if err != nil {
		return nil, err
	}

	authMetadata := NewAuthMetadataVars(params.Metadata)

	conn.accountKey = authMetadata.AccountKey

	return conn, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	return &Connector{Connector: base}, nil
}
