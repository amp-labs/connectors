package marketo

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
)

// Write creates/updates records in marketo. Write currently supports operations to the leads API only.
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := c.getAPIURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	payload, err := c.constructPayload(config)
	if err != nil {
		return nil, err
	}

	json, err := c.Client.Post(ctx, url.String(), payload)
	if err != nil {
		return nil, err
	}

	resp, err := common.UnmarshalJSON[writeResponse](json)
	if err != nil {
		return nil, err
	}

	// In case of an empty result, we return a zero-valued WriteResult.
	if len(resp.Result) == 0 {
		return &common.WriteResult{Success: true}, nil
	}

	recordId, err := constructId(config.ObjectName, resp)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  resp.Success,
		RecordId: recordId,
		Data:     resp.Result[0],
	}, nil
}

func constructId(objectName string, resp *writeResponse) (string, error) {
	var (
		recordId any
		success  bool
		err      error
	)

	switch {
	case usesStandardId(objectName):
		recordId = resp.Result[0]["id"]
		// By default the recordId is returned as a float64
		recordId, success = recordId.(float64)

		err = checkErr(resp, recordId, success)
		if err != nil {
			return "", err
		}

	case usesMarketoGUID(objectName):
		recordId = resp.Result[0]["marketoGUID"]
		// By default the marketoGUID is returned as a string
		recordId, success = recordId.(string)

		err = checkErr(resp, recordId, success)
		if err != nil {
			return "", err
		}
	case resp.Success:
		return "", nil
	default:
		return "", checkErr(resp, "", resp.Success)
	}

	return fmt.Sprint(recordId), nil
}

func checkErr(resp *writeResponse, recordId any, success bool) (err error) {
	if !success || recordId == "" || recordId == 0 {
		// This means there is a recordLevel error.
		// We construct the error and send it back to the client.
		message, err := constructErrMessage(resp.Result)
		if err != nil {
			return err
		}

		return errors.New(message) // nolint:err113
	}

	return nil
}

type payload struct {
	Action      string           `json:"action"`
	LookupField string           `json:"lookupField,omitempty"`
	Input       []map[string]any `json:"input"`
}

func (c *Connector) constructPayload(config common.WriteParams) (payload, error) {
	data, ok := config.RecordData.(map[string]any)
	if !ok {
		return payload{}, ErrInvalidData
	}

	// If we're updating leads in marketo.
	if config.ObjectName == leads && len(config.RecordId) > 0 {
		id, err := strconv.Atoi(config.RecordId)
		if err != nil {
			return payload{}, err
		}

		data["id"] = id

		return payload{
			Action:      "updateOnly",
			LookupField: "id",
			Input:       []map[string]any{data},
		}, nil
	}

	// The rest of supported objects will use this generic schema.
	return payload{
		Action: "createOrUpdate",
		Input:  []map[string]any{data},
	}, nil
}
