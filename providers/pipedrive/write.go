package pipedrive

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

type writeResponse struct {
	Data    map[string]any `json:"data"`
	Success bool           `json:"success"`
	// Other fields.
}

// Write creates or updates records in a pipedriver account.
// https://developers.pipedrive.com/docs/api/v1
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	var write common.WriteMethod

	apiVersion := apiV1
	if c.moduleID == providers.PipedriveV2 {
		apiVersion = apiV2
	}

	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := c.getAPIURL(config.ObjectName, apiVersion)
	if err != nil {
		return nil, err
	}

	if !v2SupportedObjects.Has(config.ObjectName) {
		return nil, common.ErrObjectNotSupported
	}

	if len(config.RecordId) != 0 {
		url.AddPath(config.RecordId)

		write = c.Client.Put
		if c.moduleID == providers.PipedriveV2 {
			write = c.Client.Patch
		}
	} else {
		write = c.Client.Post
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
