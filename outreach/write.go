package outreach

import (
	"context"
	"encoding/json"
	"fmt"

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
	var (
		write common.WriteMethod
		data  map[string]any
		err   error
	)

	url, err := c.getApiURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	// prepares the updating data request.
	if len(config.RecordId) > 0 {
		url.AddPath(config.RecordId)

		write = c.Client.Patch
	} else {
		// prepares the creating data request.
		write = c.Client.Post
	}

	data, err = parseData(config)
	if err != nil {
		return nil, err
	}

	res, err := write(ctx, url.String(), data, JSONAPIContentTypeHeader)
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
