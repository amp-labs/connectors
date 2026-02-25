package outreach

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

var _ connectors.BatchRecordReaderConnector = &Connector{}

// GetRecordsByIds implements BatchRecordReaderConnector for Outreach.
// It fetches records for the given object and IDs, returning a ReadResult for each.
func (c *Connector) GetRecordsByIds(ctx context.Context, params common.ReadByIdsParams) ([]common.ReadResultRow, error) {
	// Sanitize method arguments.
	config := common.ReadParams{
		ObjectName:        params.ObjectName,
		Fields:            datautils.NewSetFromList(params.Fields),
		AssociatedObjects: params.AssociatedObjects,
	}

	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadByIDsURL(params.ObjectName, params.RecordIds)
	if err != nil {
		return nil, err
	}

	// Sets the query parameter `include` when the request has associated objects.
	if len(params.AssociatedObjects) > 0 {
		url.WithQueryParam(includeQueryParam, strings.Join(params.AssociatedObjects, ","))
	}

	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	included, err := common.UnmarshalJSON[includedObjects](rsp)
	if err != nil {
		return nil, err
	}

	readResult, err := common.ParseResult(
		rsp,
		getRecords,
		getNextRecordsURL,
		getOutreachDataMarshaller(config, included.Included, common.FlattenNestedFields(attributesKey)),
		config.Fields,
	)
	if err != nil {
		return nil, err
	}

	return readResult.Data, nil
}

// buildReadByIDsURL constructs a URL for fetching multiple records by their IDs.
// Uses Outreach API's filter syntax where IDs are comma-separated: filter[id]=1,2,3
// This is more efficient than making individual requests for each ID.
func (c *Connector) buildReadByIDsURL(objectName string, ids []string) (*urlbuilder.URL, error) {
	url, err := c.getApiURL(objectName)
	if err != nil {
		return nil, err
	}

	// Join IDs with commas for the filter parameter
	// Outreach API supports: filter[id]=1,2,3
	if len(ids) > 0 {
		url.WithQueryParam("filter[id]", strings.Join(ids, ","))
	}

	return url, nil
}
