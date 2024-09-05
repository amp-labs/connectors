package marketo

import (
	"context"
	"errors"
	"fmt"

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

	json, err := c.Client.Post(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	resp, err := common.UnmarshalJSON[writeResponse](json)
	if err != nil {
		return nil, err
	}

	if len(resp.Result) == 0 {
		return nil, ErrEmptyResultResponse
	}

	recordId, err := constructId(config.ObjectName, resp)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  resp.Success,
		RecordId: fmt.Sprint(recordId),
		Data:     resp.Result[0],
	}, nil
}

func constructId(objectName string, resp *writeResponse) (string, error) {
	var (
		recordId any
		success  bool
	)

	switch {
	case usesStandardId(objectName):
		recordId = resp.Result[0]["id"]
		// By default the recordId is returned as a float64
		recordId, success = recordId.(float64)
		if !success || recordId == 0 {
			// This means, there is a recordLevel error.
			// We return the recordLevel Err to the client
			message, err := constructErrMessage(resp.Result)
			if err != nil {
				return "", err
			}

			return "", errors.New(message) //nolint: goerr113
		}

	case usesMarketoGUID(objectName):
		recordId = resp.Result[0]["marketoGUID"]
		// By default the recordId is returned as a float64
		recordId, success = recordId.(string)
		if !success || recordId == "" {
			// This means, there is a recordLevel error.
			message, err := constructErrMessage(resp.Result)
			if err != nil {
				return "", err
			}

			return "", errors.New(message) //nolint: goerr113
		}

	default:
		return "", common.ErrOperationNotSupportedForObject
	}

	return fmt.Sprint(recordId), nil
}
