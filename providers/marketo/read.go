package marketo

import (
	"context"
	"errors"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

type readResponse struct {
	Result        []leadActivity `json:"result"`
	MoreResult    bool           `json:"moreResult"`
	NextPageToken string         `json:"nextPageToken"`
}

type leadActivity struct {
	LeadID int `json:"leadId"`
	// Other fields
}

// Read retrieves data based on the provided common.ReadParams configuration parameters.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.constructReadURL(ctx, config)
	if err != nil {
		// If this is the case, we return a zero records response.
		if errors.Is(err, ErrZeroRecords) {
			return &common.ReadResult{
				Data: []common.ReadResultRow{},
				Done: true,
			}, nil
		}

		return nil, err
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res,
		getRecords,
		constructNextRecordsURL(config.ObjectName),
		common.GetMarshaledData,
		config.Fields,
	)
}

func (c *Connector) ReadLeadsByID(ctx context.Context, leadIds []string, fields datautils.StringSet) (*common.ReadResult, error) {
	url, err := c.getAPIURL(leads)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(filterTypeQuery, idFilter)
	url.WithQueryParam(filterValuesQuery, strings.Join(leadIds, ","))

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res,
		getRecords,
		constructNextRecordsURLForLeadsByID(),
		common.GetMarshaledData,
		fields,
	)
}

func constructNextRecordsURLForLeadsByID() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return jsonquery.New(node).StrWithDefault("nextPageToken", "")
	}
}
