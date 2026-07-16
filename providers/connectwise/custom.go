package connectwise

import (
	"context"
	"fmt"
	"maps"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/connectwise/internal/batch"
	"github.com/spyzhov/ajson"
)

// flattenCustomFields returns a RecordTransformer that lifts customFields
// from the nested ConnectWise response into the top-level record map.
func flattenCustomFields() common.RecordTransformer {
	return func(node *ajson.Node) (map[string]any, error) {
		customFields, err := jsonquery.New(node).ArrayOptional("customFields")
		if err != nil {
			return nil, err
		}

		root, err := jsonquery.Convertor.ObjectToMap(node)
		if err != nil {
			return nil, err
		}

		if len(customFields) == 0 {
			// This object doesn't have custom fields defined.
			// Nothing to attach. Must return raw root as is.
			return root, nil
		}

		fields := make(map[string]any)

		for _, customField := range customFields {
			field, err := jsonquery.ParseNode[readCustomField](customField)
			if err != nil {
				return nil, err
			}

			fields[field.makeFieldName()] = field.Value
		}

		// Move custom fields to the top root level.
		maps.Copy(root, fields)

		return root, nil
	}
}

// requestCustomFields fetches custom field definitions for objectName.
// It samples one record to collect custom field IDs, then loads the matching
// definitions and returns them indexed by ID.
func (c *Connector) requestCustomFields(
	ctx context.Context, objectName string,
) (map[string]modelCustomField, error) {
	if !objectsSupportingCustomFields.Has(objectName) {
		// This object cannot have custom fields, we are done.
		return map[string]modelCustomField{}, nil
	}

	customFieldIds, err := c.getCustomFieldIdsFromSampleRecord(ctx, objectName)
	if err != nil {
		return nil, err
	}

	if len(customFieldIds) == 0 {
		// This object does not have custom fields, we are done.
		return map[string]modelCustomField{}, nil
	}

	customFieldDefinitions, err := batch.Read[modelCustomField](
		ctx, c.batchAdapter, "userDefinedFields", customFieldIds)
	if err != nil {
		return nil, err
	}

	fields := make(map[string]modelCustomField)
	for _, def := range customFieldDefinitions {
		fields[strconv.Itoa(def.Id)] = def
	}

	return fields, nil
}

// objectsSupportingCustomFields lists ConnectWise object names that can include customFields in read responses.
// From the ConnectWise documentation:
// > Wondering if an endpoint supports custom fields?
// > Supported endpoints will have customFields(CustomFieldValue[])
// > listed at the end of the documentation.
var objectsSupportingCustomFields = datautils.NewSet( // nolint:gochecknoglobals
	"activities",
	"adjustments",
	"agreements",
	"calculateSla",
	"catalog",
	"companies",
	"configurations",
	"contacts",
	"expense/entries",
	"finance/companyFinance",
	"invoices",
	"orders",
	"products",
	"project/tickets",
	"projects",
	"purchaseorders",
	"rmaTags",
	"sales/opportunities",
	"service/tickets",
	"system/members",
	"system/membertemplates",
	"time/entries",
	"withSso",
)

// getCustomFieldIdsFromSampleRecord reads one record for objectName and returns
// the IDs of the custom fields present in that response.
// It is guaranteed that all custom fields will be returned even if they are empty and set to null.
func (c *Connector) getCustomFieldIdsFromSampleRecord(ctx context.Context, objectName string) ([]string, error) {
	url, err := c.getURL(objectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("pageSize", "1")

	res, err := c.JSONHTTPClient().Get(ctx, url.String(), c.clientIdHeader())
	if err != nil {
		return nil, err
	}

	records, err := common.UnmarshalJSON[readResponses](res)
	if err != nil {
		return nil, err
	}

	if len(*records) == 0 {
		// There are no records for sampling.
		return []string{}, nil
	}

	record := (*records)[0]

	ids := make([]string, len(record.CustomFields))
	for i, field := range record.CustomFields {
		ids[i] = strconv.Itoa(field.Id)
	}

	return ids, nil
}

// readResponses represents the response shape returned by object read endpoints.
// Only the customFields field is modeled here.
type readResponses []readResponse

// readResponse represents a single object read result.
type readResponse struct {
	CustomFields []readCustomField `json:"customFields"`
}

// readCustomField models the customFields item embedded in a read response.
type readCustomField struct {
	Id                    int    `json:"id"`
	Caption               string `json:"caption"`
	Type                  string `json:"type"`
	EntryMethod           string `json:"entryMethod"`
	NumberOfDecimals      int    `json:"numberOfDecimals"`
	Value                 any    `json:"value"`
	ConnectWiseId         string `json:"connectWiseId"`
	RowNum                int    `json:"rowNum"`
	UserDefinedFieldRecId int    `json:"userDefinedFieldRecId"`
	PodId                 string `json:"podId"`
}

// modelCustomField represents a custom field definition returned by ConnectWise.
// Endpoint: "/system/userDefinedFields".
type modelCustomField struct {
	Id                  int    `json:"id"`
	PodId               int    `json:"podId"`
	PodDescription      string `json:"podDescription"`
	Caption             string `json:"caption"`
	SequenceNumber      int    `json:"sequenceNumber"`
	ScreenId            string `json:"screenId"`
	ScreenDescription   string `json:"screenDescription"`
	FieldTypeIdentifier string `json:"fieldTypeIdentifier"`
	NumberDecimals      int    `json:"numberDecimals"`
	EntryTypeIdentifier string `json:"entryTypeIdentifier"`
	RequiredFlag        bool   `json:"requiredFlag"`
	DisplayOnScreenFlag bool   `json:"displayOnScreenFlag"`
	ReadOnlyFlag        bool   `json:"readOnlyFlag"`
	ListViewFlag        bool   `json:"listViewFlag"`
	Options             []struct {
		Id           int    `json:"id"`
		OptionValue  string `json:"optionValue"`
		DefaultFlag  bool   `json:"defaultFlag"`
		InactiveFlag bool   `json:"inactiveFlag"`
		SortOrder    int    `json:"sortOrder"`
	} `json:"options"`
	BusinessUnitIds     []any     `json:"businessUnitIds"`
	LocationIds         []any     `json:"locationIds"`
	ConnectWiseID       string    `json:"connectWiseID"`
	DisplayScreenInASIO bool      `json:"displayScreenInASIO"`
	DateCreated         time.Time `json:"dateCreated"`
	Info                any       `json:"_info"`
}

// getValueType returns the common.ValueType corresponding to the custom field.
func (f modelCustomField) getValueType() common.ValueType {
	switch f.FieldTypeIdentifier {
	case "Button":
		return common.ValueTypeOther
	case "Checkbox":
		return common.ValueTypeBoolean
	case "Date":
		return common.ValueTypeDate
	case "Hyperlink":
		return common.ValueTypeString
	case "TextArea":
		return common.ValueTypeString
	case "Number":
		return multiValueOrDefault(f, common.ValueTypeFloat)
	case "Percent":
		return multiValueOrDefault(f, common.ValueTypeFloat)
	case "Text":
		return multiValueOrDefault(f, common.ValueTypeString)
	default:
		return common.ValueTypeOther
	}
}

// multiValueOrDefault returns MultiSelect or SingleSelect for list-style fields, and defaultValue otherwise.
// Number, Percent, Text can all be either primitive data type or Multi/Single Select.
func multiValueOrDefault(field modelCustomField, defaultValue common.ValueType) common.ValueType {
	switch field.EntryTypeIdentifier {
	case "List":
		return common.ValueTypeMultiSelect
	case "Option":
		return common.ValueTypeSingleSelect
	default:
		return defaultValue
	}
}

func (f modelCustomField) getProviderType() string {
	return f.EntryTypeIdentifier + "_" + f.FieldTypeIdentifier
}

func (f modelCustomField) getValues() []common.FieldValue {
	values := make(common.FieldValues, len(f.Options))
	for i, option := range f.Options {
		values[i] = common.FieldValue{
			Value:        option.OptionValue,
			DisplayValue: option.OptionValue,
		}
	}

	if len(values) == 0 {
		return nil
	}

	return values
}

func (f modelCustomField) makeFieldName() string {
	return fmt.Sprintf("customField%v", strconv.Itoa(f.Id))
}

func (f readCustomField) makeFieldName() string {
	return fmt.Sprintf("customField%v", strconv.Itoa(f.Id))
}
