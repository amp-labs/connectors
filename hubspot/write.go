package hubspot

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

type writeResponse struct {
	CreatedAt             string         `json:"createdAt"`
	Archived              bool           `json:"archived"`
	ArchivedAt            string         `json:"archivedAt"`
	PropertiesWithHistory any            `json:"propertiesWithHistory"`
	ID                    string         `json:"id"`
	Properties            map[string]any `json:"properties"`
	UpdatedAt             string         `json:"updatedAt"`
}

type writeMethod func(context.Context, string, any, ...common.Header) (*common.JSONHTTPResponse, error)

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	var write writeMethod

	url := fmt.Sprintf("%s/objects/%s", c.BaseURL, config.ObjectName)

	if config.ObjectId != "" {
		write = c.Client.Patch
		url = fmt.Sprintf("%s/%s", url, config.ObjectId)
	} else {
		write = c.Client.Post
	}

	json, err := write(ctx, url, config.ObjectData)
	if err != nil {
		return nil, err
	}

	rsp, err := common.UnmarshalJSON[writeResponse](json)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		ObjectId: rsp.ID,
		Success:  true,
		Data:     rsp.Properties,
	}, nil
}
