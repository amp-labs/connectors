package custom

import (
	"sync"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

func newBatchPayload(fieldDefinitions []common.FieldDefinition) *BatchPayload {
	payloads := make([]Payload, len(fieldDefinitions))

	for index, definition := range fieldDefinitions {
		payloads[index] = *newPayload(definition)
	}

	return &BatchPayload{
		Inputs: payloads,
	}
}

func newPayload(definition common.FieldDefinition) *Payload {
	return &Payload{
		Name:            definition.FieldName,
		Label:           definition.DisplayName,
		GroupName:       definition.UniqueProperties.HubspotGroupName,
		Type:            matchType(definition),
		FieldType:       matchFieldType(definition),
		Hidden:          false,
		Description:     definition.Description,
		FormField:       true,
		DataSensitivity: "non_sensitive",
		HasUniqueValue:  definition.Unique,
		Options:         newOptions(definition),
	}
}

func matchType(definition common.FieldDefinition) string {
	switch definition.ValueType {
	case common.FieldTypeString:
		return "string"
	case common.FieldTypeBoolean:
		return "bool"
	case common.FieldTypeDate:
		return "date"
	case common.FieldTypeDateTime:
		return "datetime"
	case common.FieldTypeSingleSelect, common.FieldTypeMultiSelect:
		return "enumeration"
	case common.FieldTypeInt, common.FieldTypeFloat:
		return "number"
	default:
		return "string"
	}
}

func matchFieldType(definition common.FieldDefinition) string {
	switch definition.ValueType {
	case common.FieldTypeString:
		return "text"
	case common.FieldTypeBoolean:
		return "booleancheckbox"
	case common.FieldTypeDate:
		return "date"
	case common.FieldTypeDateTime:
		return "datetime"
	case common.FieldTypeSingleSelect:
		return "radio"
	case common.FieldTypeMultiSelect:
		return "select"
	case common.FieldTypeInt, common.FieldTypeFloat:
		return "number"
	default:
		return "text"
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
		result.populateField(fields, action, nil)
	}
}

func (r Response) populateField(
	fields map[string]common.FieldUpsertResult, action common.UpsertMetadataAction, mutex *sync.Mutex,
) {
	metadata, err := datautils.StructToMap(r)
	if err != nil {
		metadata = nil // No need to fail. Sending empty data is fine.
	}

	if mutex != nil {
		// Concurrent processing.
		mutex.Lock()
		defer mutex.Unlock()
	}

	fields[r.Name] = common.FieldUpsertResult{
		FieldName: r.Name,
		Action:    action,
		Metadata:  metadata,
		Warnings:  nil,
	}
}
