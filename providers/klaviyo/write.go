package klaviyo

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := c.getWriteURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	var write common.WriteMethod

	if len(config.RecordId) == 0 {
		if supportedObjectsByCreate[c.Module.ID].Has(config.ObjectName) {
			write = c.Client.Post
		}
	} else {
		if supportedObjectsByUpdate[c.Module.ID].Has(config.ObjectName) {
			write = c.Client.Patch

			url.AddPath(config.RecordId)
		}
	}

	if write == nil {
		// No supported REST operation was found for current object.
		return nil, common.ErrOperationNotSupportedForObject
	}

	payload, err := prepareWritePayload(config)
	if err != nil {
		return nil, errors.Join(common.ErrPreprocessingWritePayload, err)
	}

	res, err := write(ctx, url.String(), payload, c.revisionHeader())
	if err != nil {
		return nil, err
	}

	body, ok := res.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	// write response was with payload
	return constructWriteResult(body)
}

type writePayload struct {
	ID         any            `json:"id,omitempty"`
	Type       any            `json:"type"`
	Attributes map[string]any `json:"attributes"`
	Links      any            `json:"links,omitempty"`
}

func prepareWritePayload(config common.WriteParams) (any, error) {
	if config.ObjectName == objectNameImageUpload {
		// This is the only object with different structure of payload.
		// There is nothing to preprocess, no-op.
		return config.RecordData, nil
	}

	// Similar to the "read" operation data is expected to be flattened.
	// For the "write" it is the opposite. The user supplies flat fields which should be nested.
	// From API analysis it is clear that both Create/Update may have the following properties:
	// * id - required for update
	// * type
	// * attributes
	// * links - optional
	//
	// Therefore, the implementation below will put any unknown properties under "attributes".
	//
	// Additionally, update operation must include identifier not only in URL path but in payload,
	// for convenience this Klaviyo API requirement is automatically satisfied.
	object, err := convertPayloadToMap(config.RecordData)
	if err != nil {
		return nil, err
	}

	payload := writePayload{
		Attributes: make(map[string]any),
		Type:       objectNameToTypeWritePayload.Get(config.ObjectName),
	}

	if len(config.RecordId) != 0 {
		// Attach id to payload if any.
		payload.ID = config.RecordId
	}

	for key, value := range object {
		switch key {
		case "id":
			// Identifier was already automatically attached for update payload.
			// But if the user specified id within request payload it will be a no-op.
			continue
		case "type":
			// Object type is implied from the object name.
			continue
		case "links":
			payload.Links = value
		default:
			// Unknown field.
			// We imply that true intent was to supply a property under "attributes" key.
			payload.Attributes[key] = value
		}
	}

	// Payload is always sent via "data" property.
	return map[string]any{
		"data": payload,
	}, nil
}

func constructWriteResult(body *ajson.Node) (*common.WriteResult, error) {
	nested, err := jsonquery.New(body).Object("data", false)
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(nested).StrWithDefault("id", "")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(nested)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}

func convertPayloadToMap(inputPayload any) (map[string]any, error) {
	if object, ok := inputPayload.(map[string]any); ok {
		return object, nil
	}

	bytes, err := json.Marshal(inputPayload)
	if err != nil {
		return nil, err
	}

	object := make(map[string]any)
	if err = json.Unmarshal(bytes, &object); err != nil {
		return nil, err
	}

	return object, nil
}
