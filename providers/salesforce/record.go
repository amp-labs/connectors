package salesforce

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

// GetRecordsWithIds returns records matching identifiers.
func (c *Connector) GetRecordsWithIds( // nolint:revive
	ctx context.Context,
	objectName string,
	ids []string,
	fields []string,
	associations []string, // no-op, preserved to match other deep connectors
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
		common.GetMarshaledData,
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

	fmt.Println("query", query)

	url.WithQueryParam("q", query)

	return url, nil
}
