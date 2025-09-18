package keap

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedObjectsByRead[common.ModuleRoot].Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	// Pagination doesn't automatically attach query params which were used for the first page.
	// Therefore, enforce request of "custom_fields" if object is applicable.
	if objectsWithCustomFields[common.ModuleRoot].Has(config.ObjectName) {
		// Request custom fields.
		url.WithQueryParam("optional_properties", "custom_fields")
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
		makeGetRecords(config.ObjectName),
		getNextRecordsURL,
		common.MakeMarshaledDataFunc(c.attachReadCustomFields(customFields)),
		config.Fields,
	)
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Next page
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := c.getReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if version2ObjectNames.Has(params.ObjectName) {
		readURLVersion2(params, url)
	} else {
		readURLVersion1(params, url)
	}

	return url, nil
}

func readURLVersion1(params common.ReadParams, url *urlbuilder.URL) {
	url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))

	if !params.Since.IsZero() {
		url.WithQueryParam("since", datautils.Time.FormatRFC3339inUTCWithMilliseconds(params.Since))
	}
}

func readURLVersion2(params common.ReadParams, url *urlbuilder.URL) {
	// Since parameter is not applicable to objects in Module V2.
	if params.ObjectName == "contact_link_types" {
		url.WithQueryParam("pageSize", strconv.Itoa(DefaultPageSize))
	} else {
		url.WithQueryParam("page_size", strconv.Itoa(DefaultPageSize))
	}

	if !params.Since.IsZero() {
		url.WithQueryParam("filter",
			fmt.Sprintf("start_update_time==%v",
				datautils.Time.FormatRFC3339inUTCWithMilliseconds(params.Since),
			),
		)
	}

	if params.ObjectName == objectNameContactsV2 {
		url.WithQueryParam("fields", "addresses,anniversary_date,birth_date,company,contact_type,create_time,custom_fields,email_addresses,family_name,fax_numbers,given_name,id,job_title,leadsource_id, links,middle_name,notes,origin,owner_id,phone_numbers,preferred_locale,preferred_name,prefix, referral_code,score_value,social_accounts,source_type,spouse_name,suffix,tag_ids,time_zone,update_time,utm_parameters,website") // nolint:lll
	}
}

// requestCustomFields makes and API call to get model describing custom fields.
// For not applicable objects the empty mapping is returned.
// The mapping is between "custom field id" and struct containing "human-readable field name".
func (c *Connector) requestCustomFields(
	ctx context.Context, objectName string,
) (map[string]modelCustomField, error) {
	if !objectsWithCustomFields[common.ModuleRoot].Has(objectName) {
		// This object doesn't have custom fields, we are done.
		return map[string]modelCustomField{}, nil
	}

	url, err := c.getModelURL(objectName)
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

	fields := make(map[string]modelCustomField)
	for _, field := range fieldsResponse.CustomFields {
		fields[field.ID.String()] = field
	}

	return fields, nil
}

// nolint:tagliatelle
type modelCustomFieldsResponse struct {
	CustomFields []modelCustomField `json:"custom_fields"`
}

// nolint:tagliatelle
type modelCustomField struct {
	ID           naming.Text `json:"id"`
	Label        string      `json:"label"`
	Options      []any       `json:"options"`
	RecordType   string      `json:"record_type"`
	FieldType    string      `json:"field_type"`
	FieldName    string      `json:"name"`
	DefaultValue any         `json:"default_value"`
}

func (f modelCustomField) Name() string {
	if f.FieldName == "" {
		return f.Label
	}

	return f.FieldName
}

// nolint:tagliatelle
type readCustomFieldsResponse struct {
	CustomFields []readCustomField `json:"custom_fields"`
}

type readCustomField struct {
	ID      naming.Text `json:"id"`
	Content any         `json:"content"`
}
