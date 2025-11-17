package outreach

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

// GetRecordsByIds returns records matching identifiers.
// It uses Outreach API's filter syntax: filter[id]=1,2,3
func (c *Connector) GetRecordsByIds( //nolint:revive
	ctx context.Context,
	objectName string,
	ids []string,
	fields []string,
	associations []string,
) ([]common.ReadResultRow, error) {
	// Sanitize method arguments.
	config := common.ReadParams{
		ObjectName:        objectName,
		Fields:            datautils.NewSetFromList(fields),
		AssociatedObjects: associations,
	}

	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadByIDsURL(objectName, ids)
	if err != nil {
		return nil, err
	}

	// Sets the query parameter `include` when the request has associated objects.
	if len(associations) > 0 {
		url.WithQueryParam(includeQueryParam, strings.Join(associations, ","))
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
// Uses Outreach API filter syntax: filter[id]=1,2,3
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
