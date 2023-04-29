package connectors

import (
	"github.com/amp-labs/connectors/salesforce"
	"github.com/amp-labs/connectors/common"
)

func Read(spec common.ReadConfig) (common.Result, error)	{
	if spec.API == common.Salesforce {
		return salesforce.Read(spec)
	}

	return common.Result{}, common.ErrorWithStatus{
		StatusCode: 400,
		Message: "API not supported",
	}
}
