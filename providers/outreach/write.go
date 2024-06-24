package outreach

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/amp-labs/connectors/common"
)

type WriteResponse struct {
	Data map[string]any `json:"data"`
}

var JSONAPIContentTypeHeader = common.Header{ //nolint:gochecknoglobals
	Key:   "Content-Type",
	Value: "application/vnd.api+json",
}

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	var write common.WriteMethod

	URL, err := url.JoinPath(c.BaseURL, config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.RecordId) > 0 {
		URL, err = url.JoinPath(URL, config.RecordId)
		if err != nil {
			return nil, err
		}

		write = c.Client.Patch
	} else {
		write = c.Client.Post
	}

	// Outreach expects everything to be wrapped in a "data" object.
	data := make(map[string]any)
	data["data"] = config.RecordData

	res, err := write(ctx, URL, data, JSONAPIContentTypeHeader)
	if err != nil {
		return nil, err
	}

	var response WriteResponse

	err = json.Unmarshal(res.Body.Source(), &response)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: fmt.Sprint(response.Data["id"]),
		Data:     response.Data,
	}, nil
}
