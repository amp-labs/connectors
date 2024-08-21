package marketo

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

type writeResponse struct {
	Result  []map[string]any `json:"result"`
	Success bool             `json:"success"`
	Errors  []map[string]any `json:"errors"`
}

// Write creates/updates records in marketo. Write currently supports operations to the leads API only.
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	url, err := c.getApiURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	// prepares the updating data request.
	if len(config.RecordId) > 0 {
		url, err = updateURLWithID(url, config.RecordId)
		if err != nil {
			return nil, err
		}
	}

	json, err := c.Client.Post(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	resp, err := common.UnmarshalJSON[writeResponse](json)
	if err != nil {
		return nil, err
	}

	if len(resp.Result) == 0 {
		return nil, ErrEmptyResultResponse
	}

	id := resp.Result[0]["id"]

	// By default the id is returned as a float64
	id, ok := id.(float64)
	if !ok || id == 0 {
		return nil, common.ErrMissingRecordID
	}

	return &common.WriteResult{
		Success:  resp.Success,
		RecordId: fmt.Sprint(id),
		Data:     resp.Result[0],
	}, nil
}
