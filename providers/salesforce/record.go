package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

// GetRecordsByIds returns records matching identifiers.
func (c *Connector) GetRecordsByIds( // nolint:revive
	ctx context.Context,
	objectName string,
	ids []string,
	fields []string,
	associations []string,
) ([]common.ReadResultRow, error) {
	// Sanitize method arguments.
	config := recordsByIDsParams{
		ReadParams: common.ReadParams{
			ObjectName:        objectName,
			Fields:            datautils.NewSetFromList(fields),
			AssociatedObjects: associations,
		},
		RecordIdentifiers: datautils.NewSetFromList(ids),
	}

	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadByIdentifierURL(config)
	if err != nil {
		return nil, err
	}

	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	readResult, err := common.ParseResult(
		rsp,
		getRecords,
		getNextRecordsURL,
		getSalesforceDataMarshaller(config.ReadParams),
		config.Fields,
	)
	if err != nil {
		return nil, err
	}

	return readResult.Data, nil
}

type recordsByIDsParams struct {
	common.ReadParams
	RecordIdentifiers datautils.Set[string]
}

func (c *Connector) buildReadByIdentifierURL(config recordsByIDsParams) (*urlbuilder.URL, error) {
	// Requesting record identifiers using SOQL query.
	url, err := c.getRestApiURL("query")
	if err != nil {
		return nil, err
	}

	query := makeSOQL(config.ReadParams).
		WithIDs(config.RecordIdentifiers.List()).
		String()

	url.WithQueryParam("q", query)

	return url, nil
}
