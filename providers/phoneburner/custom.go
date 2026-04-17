package phoneburner

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// customFieldKeyPrefix avoids collisions between PhoneBurner custom field display names
// and built-in contact fields (e.g. "First Name" vs first_name).
const customFieldKeyPrefix = "custom_"

var nonAlphaNumUnderscore = regexp.MustCompile(`[^a-z0-9_]+`) //nolint:gochecknoglobals

// memberCustomFieldDefinition is a row from GET /rest/1/customfields (member-level definitions).
// See: https://www.phoneburner.com/developer/route_list#customfields
type memberCustomFieldDefinition struct {
	CustomFieldID string `json:"custom_field_id"`
	DisplayName   string `json:"display_name"`
	TypeID        string `json:"type_id"`
	TypeName      string `json:"type_name"`
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

// normalizeCustomFieldKey builds a stable, lower_snake_case key used in metadata and read flattening.
func normalizeCustomFieldKey(label string) string {
	s := strings.TrimSpace(strings.ToLower(label))
	s = strings.ReplaceAll(s, " ", "_")
	s = nonAlphaNumUnderscore.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")

	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}

	return s
}

func customFieldMetadataKey(displayName string) string {
	return customFieldKeyPrefix + normalizeCustomFieldKey(displayName)
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

		out = append(out, *def)
	}

	totalPages, err := jsonquery.New(wrapper).IntegerWithDefault("total_pages", 1)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %w", common.ErrResolvingCustomFields, err)
	}

	return out, int(totalPages), nil
}

// flattenContactCustomFieldsInMap promotes each contacts GET "custom_fields" entry to a top-level key
// using the same naming as ListObjectMetadata (custom_ + normalized display name / name).
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

// getMarshaledDataContactsWithCustomFieldsPreservingRaw flattens custom field values into
// ReadResultRow.Fields while leaving ReadResultRow.Raw as the provider payload (Slab:
// "ReadResultRow.Raw should not be modified").
func getMarshaledDataContactsWithCustomFieldsPreservingRaw(
	records []map[string]any, fields []string,
) ([]common.ReadResultRow, error) {
	data := make([]common.ReadResultRow, len(records))

	fields = append(fields, "id")

	//nolint:varnamelen
	for i, record := range records {
		rawSnapshot := maps.Clone(record)
		flattenContactCustomFieldsInMap(record)

		data[i] = common.ReadResultRow{
			Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
			Raw:    rawSnapshot,
		}

		var id string

		switch v := data[i].Fields["id"].(type) {
		case string:
			id = v
		case float64:
			id = strconv.FormatFloat(v, 'f', -1, 64)
		case json.Number:
			id = v.String()
		}

		data[i].Id = id
	}

	return data, nil
}
