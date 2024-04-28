package outreach

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/amp-labs/connectors/common"
)

type WriteResponse struct {
	Data struct {
		Attributes    map[string]any `json:"attributes"`
		Relationships map[string]any `json:"relationships"`
		Links         map[string]any `json:"links"`
		ID            int            `json:"id"`
		Type          string         `json:"type"`
	} `json:"data"`
}

var JSONAPIContentTypeHeader = common.Header{ //nolint:gochecknoglobals
	Key:   "Content-Type",
	Value: "application/vnd.api+json",
}

type writeMethod func(context.Context, string, any, ...common.Header) (*common.JSONHTTPResponse, error)

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	var write writeMethod

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

	// Outreach wraps everything in a "data" object.
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

	resdata := make(map[string]any)
	resdata["data"] = response.Data

	return &common.WriteResult{
		Success:  true,
		RecordId: fmt.Sprint(response.Data.ID),
		Data:     resdata,
	}, nil
}
