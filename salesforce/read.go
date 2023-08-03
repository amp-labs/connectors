package salesforce

import (
	"github.com/amp-labs/connectors/common"
)

func (s SalesforceConnector) Read(config common.ReadConfig) (*common.ReadResult, *common.ErrorWithStatus) {
	// TODO: implement
	// Construct a SOQL query out of the config, make the appropriate API calls, handle pagination, and return the full list of objects.

	// Should return a common.ErrorWithStatus if there is an error.
	return &common.ReadResult{}, nil
}
