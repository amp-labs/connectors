package zohocrm

import (
	"context"

	"github.com/amp-labs/connectors/common"
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
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
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

	// ZohoCRM requires everything to be wrapped in a "data" object.
	// RecordData should be a list of map[string]any
	body := map[string]any{
		"data": config.RecordData,
	}

	resp, err := write(ctx, url.String(), body)
	if err != nil {
		return nil, err
	}

	response, err := common.UnmarshalJSON[writeResponse](resp)
	if err != nil {
		return nil, err
	}

	var errors []any

	// Looping in the response data to see if there is
	// an error in any of the record Responses.
	for _, r := range response.Data {
		if r["code"] != "SUCCESS" {
			errors = append(errors, r)
		}
	}

	return &common.WriteResult{
		Success: true,
		Errors:  errors,
	}, nil
}