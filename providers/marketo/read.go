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
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.constructReadURL(ctx, config)
	if err != nil {
		// If this is the case, we return an zero records response.
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
