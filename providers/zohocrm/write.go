package zohocrm

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/naming"
)

/*
Sample Response Data:

{
    "data": [
        {
            "code": "SUCCESS",
            "details": {
                "Modified_Time": "2023-05-10T01:10:47-07:00",
                "Modified_By": {
                    "name": "Patricia Boyle",
                    "id": "5725767000000411001"
                },
                "Created_Time": "2023-05-10T01:10:47-07:00",
                "id": "5725767000000524157",
                "Created_By": {
                    "name": "Patricia Boyle",
                    "id": "5725767000000424162"
                },
                "$approval_state": "approved"
            },
            "message": "record added",
            "status": "success"
        },
		{...}
    ]
}
*/

type writeResponse struct {
	Data []map[string]any `json:"data"`
}

// Write creates or updates records in a zohoCRM account.
// A maximum of 100 records can be inserted per API call.
// https://www.zoho.com/crm/developer/docs/api/v6/insert-records.html
//
// nolint: funlen, cyclop
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	ctx = logging.With(ctx, "connector", "zoho CRM")

	var (
		errs     []any
		recordId string
	)

	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	var write common.WriteMethod

	// Object names in ZohoCRM API are case sensitive.
	// Capitalizing the first character of object names to form correct URL.
	obj := naming.CapitalizeFirstLetterEveryWord(config.ObjectName)

	url, err := c.getAPIURL(obj)
	if err != nil {
		return nil, err
	}

	if len(config.RecordId) != 0 {
		url.AddPath(config.RecordId)

		write = c.Client.Put
	} else {
		write = c.Client.Post
	}

	body, err := constructWritePayload(config.RecordData)
	if err != nil {
		return nil, err
	}

	resp, err := write(ctx, url.String(), body)
	if err != nil {
		return nil, err
	}

	response, err := common.UnmarshalJSON[writeResponse](resp)
	if err != nil {
		return nil, err
	}

	// Validate that the response contains data before processing.
	// The Zoho CRM API should always return response data for write operations.
	if len(response.Data) < 1 {
		logging.Logger(ctx).Error("failed to retrieve the created/updated response data", "object", config.ObjectName,
			"response", response)

		return &common.WriteResult{Success: true}, nil
	}

	// Note: Currently handles single-record operations only (non-bulk)
	for _, r := range response.Data {
		if r["code"] != "SUCCESS" {
			errs = append(errs, r)
		}

		details, ok := r["details"].(map[string]any)
		if !ok {
			logging.Logger(ctx).Error("failed to retrieve details in response data", "object", config.ObjectName,
				"response", response)

			return &common.WriteResult{Success: true}, nil
		}

		// Extract record ID from successful responses
		if id, ok := details["id"].(string); ok && id != "" {
			recordId = id
		} else {
			logging.Logger(ctx).Error("failed to construct recordId from response",
				"object", config.ObjectName,
				"response", response)
		}
	}

	return &common.WriteResult{
		Success:  true,
		Errors:   errs,
		RecordId: recordId,
		Data:     response.Data[0],
	}, nil
}

func constructWritePayload(payload any) (any, error) {
	data, ok := payload.([]map[string]any)
	if !ok {
		objectData, ok := payload.(map[string]any)
		if !ok {
			return nil, common.ErrBadRequest
		}

		capitalizeKeys(objectData)

		return map[string]any{"data": []map[string]any{objectData}}, nil
	}

	// Range Over the Slice for every map, Capitalize them.
	for _, v := range data {
		capitalizeKeys(v)
	}

	return map[string]any{"data": data}, nil
}

func capitalizeKeys(data map[string]any) {
	// Capitalize words in the data fields for Creation/Updating
	for k, d := range data {
		fld := constructFieldNames([]string{k})
		data[fld] = d
		// Remove the previous field key in the map, as it's no longer required.
		if fld != k {
			delete(data, k)
		}
	}
}
