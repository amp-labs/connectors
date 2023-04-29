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

func Read(api API, config ReadConfig) (Result, error)	{
	if api == Salesforce {
		return salesforce.Read(config)
	}

	return Result{}, ErrorWithStatus{
		StatusCode: 400,
		Message: "API not supported",
	}
}
