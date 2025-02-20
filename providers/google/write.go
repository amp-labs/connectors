package google

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) { // nolint: cyclop,lll
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := c.getURL(config.ObjectName)
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
			write = c.Client.Put

			url.AddPath(config.RecordId)
		}
	}

	if write == nil {
		// No supported REST operation was found for current object.
		return nil, common.ErrOperationNotSupportedForObject
	}

	if c.Module.ID == ModuleCalendar {
		if config.ObjectName == objectNameCalendarList {
			if needsCalendarColorQueryParam(config) {
				url.WithQueryParam("colorRgbFormat", "true")
			}
		}
	}

	if c.Module.ID == ModulePeople {

	}

	res, err := write(ctx, url.String(), config.RecordData)
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

func constructWriteResult(node *ajson.Node) (*common.WriteResult, error) {
	resourceName, err := jsonquery.New(node).StringRequired("resourceName")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	_, recordID, ok := resourceIdentifierFormat(resourceName)
	if !ok {
		return &common.WriteResult{
			Success:  true,
			RecordId: "",
			Errors:   []any{errors.New("cannot infer record identifier")},
			Data:     data,
		}, nil
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}

// When either background or foreground is specified this means
// we must attach a query parameter for the Write request to succeed.
//
// https://developers.google.com/calendar/api/v3/reference/calendarList/insert
func needsCalendarColorQueryParam(config common.WriteParams) bool {
	properties, err := common.RecordDataToMap(config.RecordData)
	if err != nil {
		return false
	}

	triggerFields := []string{
		"foregroundColor", "backgroundColor",
	}

	for _, field := range triggerFields {
		if _, ok := properties[field]; ok {
			return true
		}
	}

	return false
}
