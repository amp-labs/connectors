package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

// GetRecordsByIds returns records matching identifiers.
//
//nolint:revive
func (c *Connector) GetRecordsByIds(ctx context.Context,
	params common.ReadByIdsParams,
) ([]common.ReadResultRow, error) {
	// Sanitize method arguments.
	config := recordsByIDsParams{
		ReadParams: common.ReadParams{
			ObjectName:        params.ObjectName,
			Fields:            datautils.NewSetFromList(params.Fields),
			AssociatedObjects: params.AssociatedObjects,
		},
		RecordIdentifiers: datautils.NewSetFromList(params.RecordIds),
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
		core.GetRecords,
		core.GetNextRecordsURL,
		core.GetDataMarshallerForRead(config.ReadParams),
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
	// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_query.htm
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
