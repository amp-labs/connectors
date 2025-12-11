package constantcontact

import (
	"context"
	"errors"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedObjectsByRead[c.Module.ID].Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	customFields, err := c.requestCustomFields(ctx, config.ObjectName)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res,
		makeGetRecords(c.Module.ID, config.ObjectName),
		makeNextRecordsURL(c.BaseURL),
		common.MakeMarshaledDataFunc(c.attachReadCustomFields(customFields)),
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		// Cursor query parameter is base64 encoded data which preserves all query parameters from initial request.
		// Therefore, this URL is ready for usage as is.
		// Example:
		// https://api.cc.email/v3/contacts?
		//			cursor=bGltaXQ9MSZuZXh0PTImdXBkYXRlZF9hZnRlcj0yMDIyLTAzLTExVDIyJTNBMDklM0EwMiUyQjAwJTNBMDA=
		return urlbuilder.New(config.NextPage.String())
	}

	// First page
	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))

	if !config.Since.IsZero() {
		switch config.ObjectName {
		case objectNameEmailCampaigns:
			sinceValue := datautils.Time.FormatRFC3339inUTC(config.Since)
			url.WithQueryParam("after_date", sinceValue)
		case objectNameContacts:
			sinceValue := datautils.Time.FormatRFC3339inUTC(config.Since)
			url.WithQueryParam("updated_after", sinceValue)
		}
	}

	if objectsWithCustomFields.Has(config.ObjectName) {
		// Request custom fields.
		url.WithQueryParam("include", "custom_fields")
	}

	return url, nil
}

// requestCustomFields makes and API call to get model describing custom fields.
// For not applicable objects the empty mapping is returned.
// The mapping is between "custom field id" and struct containing "human-readable field name".
func (c *Connector) requestCustomFields(
	ctx context.Context, objectName string,
) (map[string]modelCustomField, error) {
	if !objectsWithCustomFields.Has(objectName) {
		// This object doesn't have custom fields, we are done.
		return map[string]modelCustomField{}, nil
	}

	// Only contacts resource supports custom fields.
	url, err := c.getURL("contact_custom_fields")
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	fieldsResponse, err := common.UnmarshalJSON[modelCustomFieldsResponse](res)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	if fieldsResponse == nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, common.ErrEmptyJSONHTTPResponse)
	}

	fields := make(map[string]modelCustomField)
	for _, field := range fieldsResponse.CustomFields {
		fields[field.ID] = field
	}

	return fields, nil
}

// nolint:tagliatelle
type modelCustomFieldsResponse struct {
	CustomFields []modelCustomField `json:"custom_fields"`
}

// nolint:tagliatelle
type modelCustomField struct {
	ID        string `json:"custom_field_id"`
	Label     string `json:"label"`
	FieldName string `json:"name"`
	FieldType string `json:"type"`
}

// nolint:tagliatelle
type readCustomFieldsResponse struct {
	CustomFields []readCustomField `json:"custom_fields"`
}

// nolint:tagliatelle
type readCustomField struct {
	ID    string `json:"custom_field_id"`
	Value any    `json:"value"`
}
