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
type ReadResult = common.ReadResult
type ErrorWithStatus = common.ErrorWithStatus
type GetCallConfig = common.GetCallConfig
type GenericResult = common.GenericResult

type Connector interface {
	MakeGetCall(config GetCallConfig) (*GenericResult, *ErrorWithStatus)
	Read(config ReadConfig) (*ReadResult, *ErrorWithStatus)
}

func NewConnector(api API, workspaceRef string, accessToken string) Connector {
	switch api {
	case Salesforce:
		return salesforce.NewConnector(workspaceRef, accessToken)
	default:
		return nil
	}
}
