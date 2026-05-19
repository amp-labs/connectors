package phoneburner

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// customFieldKeyPrefix is a connector-only namespace (not in PhoneBurner JSON). The provider
// uses display_name on definitions and name on each contact custom_fields entry; we prefix those
// labels so flattened values do not collide with built-in contact keys.
const customFieldKeyPrefix = "custom_"

// memberCustomFieldDefinition is a row from GET /rest/1/customfields (member-level definitions).
// See: https://www.phoneburner.com/developer/route_list#customfields
type memberCustomFieldDefinition struct {
	CustomFieldID string `json:"custom_field_id"`
	DisplayName   string `json:"display_name"`
	TypeID        string `json:"type_id"`
	TypeName      string `json:"type_name"`
}

// isUsableForMetadata reports whether a row from GET /rest/1/customfields can be
// exposed on contacts metadata and read flattening.
//
// PhoneBurner documents display_name as required when creating a field (POST) and
// always includes it in GET examples (https://www.phoneburner.com/developer/route_list#customfields).
// The list endpoint can still return individual rows with a blank display_name (legacy or
// partial records); we skip those because our connector keys fields as custom_<display_name>,
// which must match the "name" on each contact's custom_fields array.
func (d memberCustomFieldDefinition) isUsableForMetadata() bool {
	return strings.TrimSpace(d.DisplayName) != ""
}

func memberCustomFieldTypeToValueType(typeID string) common.ValueType {
	switch strings.TrimSpace(typeID) {
	case "1":
		return common.ValueTypeString
	case "2":
		return common.ValueTypeBoolean
	case "3":
		return common.ValueTypeDate
	case "6":
		return common.ValueTypeSingleSelect
	case "7":
		return common.ValueTypeFloat
	default:
		return common.ValueTypeOther
	}
}

// customFieldMetadataKey builds the ListObjectMetadata / read field name from the provider label
// (display_name or custom_fields[].name). Callers pass the API string; do not assemble custom_* keys by hand.
// common.ExtractLowercaseFieldsFromRaw lowercases keys on ReadResultRow.Fields.
func customFieldMetadataKey(displayName string) string {
	return customFieldKeyPrefix + strings.TrimSpace(displayName)
}

func (c *Connector) fetchMemberCustomFieldDefinitions(ctx context.Context) ([]memberCustomFieldDefinition, error) {
	var out []memberCustomFieldDefinition

	page := 1

	for {
		defs, totalPages, err := c.getMemberCustomFieldDefinitionsPage(ctx, page)
		if err != nil {
			return nil, err
		}

		out = append(out, defs...)

		if page >= totalPages {
			break
		}

		page++
	}

	return out, nil
}

func (c *Connector) getMemberCustomFieldDefinitionsPage(ctx context.Context, page int) (
	[]memberCustomFieldDefinition, int, error,
) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restPrefix, restVer, "customfields")
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %w", common.ErrResolvingCustomFields, err)
	}

	url.WithQueryParam("page_size", "100")
	url.WithQueryParam("page", strconv.Itoa(page))

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %w", common.ErrResolvingCustomFields, err)
	}

	if err := interpretPhoneBurnerEnvelopeError(resp); err != nil {
		return nil, 0, fmt.Errorf("%w: %w", common.ErrResolvingCustomFields, err)
	}

	body, ok := resp.Body()
	if !ok {
		return nil, 0, fmt.Errorf("%w: empty customfields body", common.ErrResolvingCustomFields)
	}

	return parseMemberCustomFieldDefinitionsPage(body)
}

func parseMemberCustomFieldDefinitionsPage(body *ajson.Node) ([]memberCustomFieldDefinition, int, error) {
	wrapper, err := jsonquery.New(body).ObjectRequired("customfields")
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %w", common.ErrResolvingCustomFields, err)
	}

	arr, err := jsonquery.New(wrapper).ArrayOptional("customfields")
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %w", common.ErrResolvingCustomFields, err)
	}

	out := make([]memberCustomFieldDefinition, 0, len(arr))

	for _, n := range arr {
		def, err := jsonquery.ParseNode[memberCustomFieldDefinition](n)
		if err != nil {
			return nil, 0, fmt.Errorf("%w: %w", common.ErrResolvingCustomFields, err)
		}

		if !def.isUsableForMetadata() {
			continue
		}

		out = append(out, *def)
	}

	totalPages, err := jsonquery.New(wrapper).IntegerWithDefault("total_pages", 1)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %w", common.ErrResolvingCustomFields, err)
	}

	return out, int(totalPages), nil
}

// flattenContactCustomFieldsInMap promotes contact custom_fields entries to top-level custom_*
// keys for common.ExtractLowercaseFieldsFromRaw. Used only from readContactRecordTransformer;
// readhelper.MakeMarshaledDataFuncWithId keeps provider-shaped Raw (separate ObjectToMap) and
// extracts ReadResultRow.Id from raw.
func flattenContactCustomFieldsInMap(record map[string]any) map[string]any {
	raw, ok := record["custom_fields"]
	if !ok || raw == nil {
		return record
	}

	list, ok := raw.([]any)
	if !ok || len(list) == 0 {
		delete(record, "custom_fields")

		return record
	}

	for _, item := range list {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}

		name, _ := obj["name"].(string)
		if name == "" {
			continue
		}

		key := customFieldMetadataKey(name)
		if v, exists := obj["value"]; exists {
			record[key] = v
		}
	}

	delete(record, "custom_fields")

	return record
}

// readContactRecordTransformer is the common.RecordTransformer passed to
// readhelper.MakeMarshaledDataFuncWithId for contacts; it only shapes the map used for Fields.
func readContactRecordTransformer(node *ajson.Node) (map[string]any, error) {
	record, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	flattenContactCustomFieldsInMap(record)

	return record, nil
}
