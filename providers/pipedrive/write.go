package pipedrive

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

type writeResponse struct {
	Data    map[string]any `json:"data"`
	Success bool           `json:"success"`
	// Other fields.
}

// Write creates or updates records in a pipedriver account.
// https://developers.pipedrive.com/docs/api/v1
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	var write common.WriteMethod

	url, err := c.getAPIURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.RecordId) != 0 {
		url.AddPath(config.RecordId)

		write = c.JSONHTTPClient().Put
	} else {
		write = c.JSONHTTPClient().Post
	}

	resp, err := write(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	response, err := common.UnmarshalJSON[writeResponse](resp)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  response.Success,
		RecordId: fmt.Sprint(response.Data["id"]),
		Data:     response.Data,
	}, nil
}
