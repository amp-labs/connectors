package marketo

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
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
func (c *Connector) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, nextPageToken, err := c.constructReadURL(ctx, params)
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
		constructNextRecordsURL(params.ObjectName, nextPageToken),
		common.GetMarshaledData,
		params.Fields,
	)
}
