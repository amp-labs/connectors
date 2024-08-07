package outreach

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

type WriteResponse struct {
	Data map[string]any `json:"data"`
}

const (
	attributesKey    string = "attributes"
	relationshipsKey string = "relationships"
	idKey            string = "id"
	typeKey          string = "type"
	dataKey          string = "data"
)

var JSONAPIContentTypeHeader = common.Header{ //nolint:gochecknoglobals
	Key:   "Content-Type",
	Value: "application/vnd.api+json",
}

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	var write common.WriteMethod

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

	req, err := constructWriteRequest(config)
	if err != nil {
		return nil, err
	}

	res, err := write(ctx, url.String(), req, JSONAPIContentTypeHeader)
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

// constructWriteRequest creates a Write request that is expected by the Outreach API.
func constructWriteRequest(cfg common.WriteParams) (map[string]any, error) {
	var (
		nestedFields = make(map[string]any)
		attributes   = make(map[string]any)
		reqData      = make(map[string]any)
	)

	// Updating requires the id in the request body.
	// Re-adding it to the request.
	if len(cfg.RecordId) > 0 {
		iD, err := strconv.Atoi(cfg.RecordId)
		if err != nil {
			return nil, ErrIdMustInt
		}

		nestedFields[idKey] = iD
	}

	received, ok := cfg.RecordData.(map[string]any) //nolint: varnamelen
	if !ok {
		return nil, common.ErrRecordDataNotJSON
	}

	// If Relationships key has data, add it on the request.
	value, ok := received[relationshipsKey]
	if ok {
		nestedFields[relationshipsKey] = value
	}

	// Adds attributes key values.
	for k, v := range received {
		if k != relationshipsKey && k != typeKey {
			attributes[k] = v
		}
	}

	// If no type provided, provides a type which should be a singular word of the ObjectName
	// is added.
	_, ok = received[typeKey]
	if !ok {
		objectType := naming.NewSingularString(cfg.ObjectName)
		received[typeKey] = objectType
	}

	nestedFields[attributesKey] = attributes
	nestedFields[typeKey] = received[typeKey]
	reqData[dataKey] = nestedFields

	return reqData, nil
}
