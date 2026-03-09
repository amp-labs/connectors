package servicedeskplus

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

var ErrUnexpectedResponse = errors.New("response data not as expected")

func (a *Adapter) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	var responseBody []byte

	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := a.getAPIURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	body, err := constructInput(config.ObjectName, config.RecordData)
	if err != nil {
		return nil, err
	}

	headers := []common.Header{
		{
			Key:   "Accept",
			Value: "application/vnd.manageengine.sdp.v3+json",
		},
		common.HeaderFormURLEncoded,
	}

	if config.RecordId != "" {
		url.AddPath(config.RecordId)

		_, responseBody, err = a.Client.HTTPClient.Put(ctx, url.String(), body, headers...)
		if err != nil {
			return nil, err
		}
	} else {
		_, responseBody, err = a.Client.HTTPClient.Post(ctx, url.String(), body, headers...)
		if err != nil {
			return nil, err
		}
	}

	recordId, record, err := constructWriteResponse(config.ObjectName, responseBody)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		Errors:   nil,
		RecordId: recordId,
		Data:     record,
	}, nil
}

func constructInput(objectName string, data any) ([]byte, error) {
	singularObjectName := naming.NewSingularString(objectName).String()

	recordData := map[string]any{
		singularObjectName: data,
	}

	jsonData, err := json.Marshal(recordData)
	if err != nil {
		return nil, err
	}

	formData := fmt.Sprintf("input_data=%s", jsonData)
	body := bytes.NewBufferString(formData)

	return body.Bytes(), nil
}

func constructWriteResponse(objectName string, response []byte) (string, map[string]any, error) {
	var responseData map[string]any
	if err := json.Unmarshal(response, &responseData); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	singularObjectName := naming.NewSingularString(objectName).String()

	record, ok := responseData[singularObjectName].(map[string]any)
	if !ok {
		return "", nil, ErrUnexpectedResponse
	}

	recordId := ""
	if id, exists := record[idField].(string); exists {
		recordId = id
	}

	return recordId, record, nil
}
