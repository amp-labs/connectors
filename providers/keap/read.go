package keap

import (
	"context"
	"errors"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/keap/metadata"
)

var ErrResolvingCustomFields = errors.New("cannot resolve custom fields")

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

	// Pagination doesn't automatically attach query params which were used for the first page.
	// Therefore, enforce request of "custom_fields" if object is applicable.
	if objectsWithCustomFields[c.Module.ID].Has(config.ObjectName) {
		// Request custom fields.
		url.WithQueryParam("optional_properties", "custom_fields")
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module.ID, config.ObjectName)

	return common.ParseResult(res,
		c.parseReadRecords(ctx, config, responseFieldName),
		makeNextRecordsURL(c.Module.ID),
		common.GetMarshaledData,
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		return urlbuilder.New(config.NextPage.String())
	}

	// First page
	url, err := c.getReadURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if c.Module.ID == ModuleV1 {
		url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))

		if !config.Since.IsZero() {
			url.WithQueryParam("since", datautils.Time.FormatRFC3339inUTCWithMilliseconds(config.Since))
		}
	} else if c.Module.ID == ModuleV2 {
		// Since parameter is not applicable to objects in Module V2.
		if config.ObjectName == "contact_link_types" {
			url.WithQueryParam("pageSize", strconv.Itoa(DefaultPageSize))
		} else {
			url.WithQueryParam("page_size", strconv.Itoa(DefaultPageSize))
		}
	}

	return url, nil
}

// requestCustomFields makes and API call to get model describing custom fields.
// For not applicable objects the empty mapping is returned.
// The mapping is between "custom field id" and struct containing "human-readable field name".
func (c *Connector) requestCustomFields(
	ctx context.Context, objectName string,
) (map[int]modelCustomField, error) {
	if !objectsWithCustomFields[c.Module.ID].Has(objectName) {
		// This object doesn't have custom fields, we are done.
		return map[int]modelCustomField{}, nil
	}

	modulePath := metadata.Schemas.LookupModuleURLPath(c.Module.ID)

	url, err := c.getURL(modulePath, objectName, "model")
	if err != nil {
		return nil, errors.Join(ErrResolvingCustomFields, err)
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, errors.Join(ErrResolvingCustomFields, err)
	}

	fieldsResponse, err := common.UnmarshalJSON[modelCustomFieldsResponse](res)
	if err != nil {
		return nil, errors.Join(ErrResolvingCustomFields, err)
	}

	fields := make(map[int]modelCustomField)
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
	ID           int    `json:"id"`
	Label        string `json:"label"`
	Options      []any  `json:"options"`
	RecordType   string `json:"record_type"`
	FieldType    string `json:"field_type"`
	FieldName    string `json:"field_name"`
	DefaultValue any    `json:"default_value"`
}

// nolint:tagliatelle
type readCustomFieldsResponse struct {
	CustomFields []readCustomField `json:"custom_fields"`
}

type readCustomField struct {
	ID      int `json:"id"`
	Content any `json:"content"`
}
