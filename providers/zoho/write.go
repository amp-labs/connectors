package zoho

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers"
	"github.com/spyzhov/ajson"
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

const dataKey = "data"

// Write creates or updates records in a zohoCRM account.
// A maximum of 100 records can be inserted per API call.
// https://www.zoho.com/crm/developer/docs/api/v6/insert-records.html
//
// nolint: funlen, cyclop
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	ctx = logging.With(ctx, "connector", "zoho CRM", "module", c.moduleID)

	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	switch c.moduleID { // nolint: exhaustive
	case providers.ModuleZohoDesk:
		return c.writeDesk(ctx, config)
	default:
		return c.writeCRM(ctx, config)
	}
}

func (c *Connector) writeDesk(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	var write common.WriteMethod

	url, err := c.getAPIURL(deskAPIVersion, config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.RecordId) != 0 {
		url.AddPath(config.RecordId)

		write = c.Client.Patch
	} else {
		write = c.Client.Post
	}

	resp, err := write(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	node, ok := resp.Body()
	if !ok {
		logging.Logger(ctx).Error("failed to retrieve the created/updated response data", "object", config.ObjectName)

		return &common.WriteResult{Success: true}, nil
	}

	recordId, err := jsonquery.New(node).StringOptional("id")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		Errors:   nil,
		RecordId: *recordId,
		Data:     data,
	}, nil
}

func (c *Connector) writeCRM(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	var (
		errs  []any
		write common.WriteMethod
	)

	// Object names in ZohoCRM API are case sensitive.
	// Capitalizing the first character of object names to form correct URL.
	obj := naming.CapitalizeFirstLetterEveryWord(config.ObjectName)

	url, err := c.getAPIURL(crmAPIVersion, obj)
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

	node, ok := resp.Body()
	if !ok {
		logging.Logger(ctx).Error("failed to retrieve the created/updated response data", "object", config.ObjectName)

		return &common.WriteResult{Success: true}, nil
	}

	records, err := jsonquery.New(node).ArrayOptional(dataKey)
	if err != nil {
		return nil, err
	}

	id, data, err := constructResponse(records, errs)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		Errors:   errs,
		RecordId: id,
		Data:     data,
	}, nil
}

func constructWritePayload(payload any) (any, error) {
	data, ok := payload.([]map[string]any)
	if !ok {
		objectData, ok := payload.(map[string]any)
		if !ok {
			return nil, common.ErrBadRequest
		}

		return map[string]any{"data": []map[string]any{objectData}}, nil
	}

	return map[string]any{"data": data}, nil
}

func constructResponse(records []*ajson.Node, errs []any) (string, map[string]any, error) {
	var (
		recordId   string
		recordData map[string]any
		err        error
	)

	for _, record := range records {
		recordData, err = jsonquery.Convertor.ObjectToMap(record)
		if err != nil {
			return "", nil, err
		}

		objectId, err := jsonquery.New(record, "details").StrWithDefault("id", "")
		if err != nil {
			return "", nil, err
		}

		code, err := jsonquery.New(record).StrWithDefault("code", "")
		if err != nil {
			return "", nil, err
		}

		if code != "SUCCESS" {
			errs = append(errs, record)
		}

		recordId = objectId
	}

	return recordId, recordData, err
}
