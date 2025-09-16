package custom

import (
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

func newBatchPayload(groupName string, fieldDefinitions []common.FieldDefinition) (*BatchPayload, error) {
	payloads := make([]Payload, len(fieldDefinitions))

	for index, definition := range fieldDefinitions {
		payload, err := newPayload(groupName, definition)
		if err != nil {
			return nil, err
		}

		payloads[index] = *payload
	}

	return &BatchPayload{
		Inputs: payloads,
	}, nil
}

func newPayload(groupName string, definition common.FieldDefinition) (*Payload, error) {
	theType, err := matchType(definition)
	if err != nil {
		return nil, err
	}

	fieldType, err := matchFieldType(definition)
	if err != nil {
		return nil, err
	}

	return &Payload{
		Name:            definition.FieldName,
		Label:           definition.DisplayName,
		GroupName:       groupName,
		Type:            theType,
		FieldType:       fieldType,
		Hidden:          false,
		Description:     definition.Description,
		FormField:       true,
		DataSensitivity: "non_sensitive",
		HasUniqueValue:  definition.Unique,
		Options:         newOptions(definition),
	}, nil
}

func matchType(definition common.FieldDefinition) (string, error) {
	switch definition.ValueType {
	case common.FieldTypeString:
		return "string", nil
	case common.FieldTypeBoolean:
		return "bool", nil
	case common.FieldTypeDate:
		return "date", nil
	case common.FieldTypeDateTime:
		return "datetime", nil
	case common.FieldTypeSingleSelect, common.FieldTypeMultiSelect:
		return "enumeration", nil
	case common.FieldTypeInt, common.FieldTypeFloat:
		return "number", nil
	default:
		return "", fmt.Errorf("%w, fieldName: %v", common.ErrFieldTypeUnknown, definition.FieldName)
	}
}

func matchFieldType(definition common.FieldDefinition) (string, error) {
	switch definition.ValueType {
	case common.FieldTypeString:
		return "text", nil
	case common.FieldTypeBoolean:
		return "booleancheckbox", nil
	case common.FieldTypeDate:
		return "date", nil
	case common.FieldTypeDateTime:
		return "datetime", nil
	case common.FieldTypeSingleSelect:
		return "radio", nil
	case common.FieldTypeMultiSelect:
		return "select", nil
	case common.FieldTypeInt, common.FieldTypeFloat:
		return "number", nil
	default:
		return "", fmt.Errorf("%w, fieldName: %v", common.ErrFieldTypeUnknown, definition.FieldName)
	}
}

type BatchPayload struct {
	Inputs []Payload `json:"inputs"`
}

type Payload struct {
	Name            string   `json:"name,omitempty"`
	Label           string   `json:"label,omitempty"`
	GroupName       string   `json:"groupName,omitempty"`
	Type            string   `json:"type,omitempty"`
	FieldType       string   `json:"fieldType,omitempty"`
	Hidden          bool     `json:"hidden,omitempty"`
	Description     string   `json:"description,omitempty"`
	FormField       bool     `json:"formField,omitempty"`
	DataSensitivity string   `json:"dataSensitivity,omitempty"`
	HasUniqueValue  bool     `json:"hasUniqueValue,omitempty"`
	Options         []Option `json:"options,omitempty"`
}

type Option struct {
	Label        string `json:"label,omitempty"`
	Value        string `json:"value,omitempty"`
	Description  string `json:"description,omitempty"`
	DisplayOrder int    `json:"displayOrder,omitempty"`
	Hidden       bool   `json:"hidden,omitempty"`
}

func newOptions(definition common.FieldDefinition) []Option {
	if definition.ValueType == common.ValueTypeBoolean {
		return []Option{{
			Label:        "True",
			Value:        "true",
			Description:  "True",
			DisplayOrder: 0,
			Hidden:       false,
		}, {
			Label:        "False",
			Value:        "false",
			Description:  "False",
			DisplayOrder: 1,
			Hidden:       false,
		}}
	}

	if definition.StringOptions == nil || len(definition.StringOptions.Values) == 0 {
		return nil
	}

	options := make([]Option, len(definition.StringOptions.Values))
	for index, value := range definition.StringOptions.Values {
		options[index] = Option{
			Label:        value,
			Value:        value,
			Description:  value,
			DisplayOrder: index,
			Hidden:       false,
		}
	}

	return options
}

type BatchResponse struct {
	CompletedAt time.Time    `json:"completedAt"`
	Status      string       `json:"status"`
	StartedAt   time.Time    `json:"startedAt"`
	Results     BatchResults `json:"results"`
	Errors      []struct {
		Status      string `json:"status"`
		Category    string `json:"category"`
		SubCategory string `json:"subCategory"`
		Message     string `json:"message"`
		Context     struct {
			Name []string `json:"name"`
		} `json:"context"`
	} `json:"errors"`
	NumErrors int `json:"numErrors"`
}

type BatchResults []Response

type Response struct {
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	Name        string    `json:"name"`
	Label       string    `json:"label"`
	Type        string    `json:"type"`
	FieldType   string    `json:"fieldType"`
	Description string    `json:"description"`
	GroupName   string    `json:"groupName"`
	Options     []struct {
		Label        string `json:"label"`
		Value        string `json:"value"`
		Description  string `json:"description"`
		DisplayOrder int    `json:"displayOrder"`
		Hidden       bool   `json:"hidden"`
	} `json:"options"`
	CreatedUserId        string `json:"createdUserId"`
	UpdatedUserId        string `json:"updatedUserId"`
	DisplayOrder         int    `json:"displayOrder"`
	Calculated           bool   `json:"calculated"`
	ExternalOptions      bool   `json:"externalOptions"`
	Archived             bool   `json:"archived"`
	HasUniqueValue       bool   `json:"hasUniqueValue"`
	Hidden               bool   `json:"hidden"`
	ModificationMetadata struct {
		Archivable         bool `json:"archivable"`
		ReadOnlyDefinition bool `json:"readOnlyDefinition"`
		ReadOnlyValue      bool `json:"readOnlyValue"`
	} `json:"modificationMetadata"`
	FormField       bool   `json:"formField"`
	DataSensitivity string `json:"dataSensitivity"`
}

// populateFields inserts FieldUpsertResult entries into the provided map, using data from each server response.
func (r BatchResults) populateFields(
	fields map[string]common.FieldUpsertResult, action common.UpsertMetadataAction,
) {
	for _, result := range r {
		result.populateField(fields, action)
	}
}

func (r Response) populateField(fields map[string]common.FieldUpsertResult, action common.UpsertMetadataAction) {
	metadata, err := datautils.StructToMap(r)
	if err != nil {
		metadata = nil // No need to fail. Sending empty data is fine.
	}

	fields[r.Name] = common.FieldUpsertResult{
		FieldName: r.Name,
		Action:    action,
		Metadata:  metadata,
		Warnings:  nil,
	}
}
