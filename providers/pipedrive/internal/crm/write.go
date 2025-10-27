package crm

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

// https://developers.pipedrive.com/docs/api/v2
func (a *Adapter) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	var write common.WriteMethod

	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := a.getAPIURL(readAPIVersion, config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.RecordId) != 0 {
		url.AddPath(config.RecordId)
		write = a.Client.Patch
	} else {
		write = a.Client.Post
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
