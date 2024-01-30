package hubspot

import (
	"context"
	"fmt"
	"strings"

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

	relativeUrl := strings.Join([]string{"objects", config.ObjectName}, "/")
	url := c.getUrl(relativeUrl)

	if config.RecordId != "" {
		write = c.Client.Patch
		url = fmt.Sprintf("%s/%s", url, config.RecordId)
	} else {
		write = c.Client.Post
	}

	fmt.Println("url", url)

	// Hubspot requires everything to be wrapped in a "properties" object.
	// We do this automatically in the write method so that the user doesn't
	// have to worry about it.
	data := make(map[string]interface{})
	data["properties"] = config.RecordData

	json, err := write(ctx, url, data)
	if err != nil {
		return nil, err
	}

	rsp, err := common.UnmarshalJSON[writeResponse](json)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		RecordId: rsp.ID,
		Success:  true,
		Data:     rsp.Properties,
	}, nil
}
