package connectors

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/salesforce"
)

type API string

const (
	Salesforce API = "salesforce"
)

// We re-export the following types so that they can be used by consumers of this library.
type ReadConfig = common.ReadConfig
type Result = common.Result
type ErrorWithStatus = common.ErrorWithStatus
type GetCallConfig = common.GetCallConfig
type GenericResult = common.GenericResult

type Connector interface {
	MakeGetCall(config GetCallConfig) (*GenericResult, error)
}

func NewConnector(api API, workspaceRef string, accessToken string) Connector {
	if api == Salesforce {
		return salesforce.NewConnector(workspaceRef, accessToken)
	}
	return nil
}
