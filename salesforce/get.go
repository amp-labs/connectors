package salesforce

import (
	"strings"

	"github.com/amp-labs/connectors/common"
)

func parseErrorMode(message string) common.ErrorMode {
	if strings.Contains(message, "INVALID_SESSION_ID") {
		return common.AccessTokenInvalid
	}
	return common.OtherError
}

func (s SalesforceConnector) MakeGetCall(c common.GetCallConfig) (*common.GenericResult, *common.ErrorWithStatus) {
	d, err := common.DoHttpGetCall(s.Client, s.BaseURL, c.Endpoint, s.AccessToken)
	if err != nil {
		if err.Mode == "" {
			err.Mode = parseErrorMode(err.Message)
		}
		return nil, err
	}
	return &common.GenericResult{Data: d}, nil
}
