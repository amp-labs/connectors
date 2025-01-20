package hubspot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
)

type objectMetadataResult struct {
	ObjectName string
	Response   common.ObjectMetadata
}

type objectMetadataError struct {
	ObjectName string
	Error      error
}

// ListObjectMetadata returns object metadata for each object name provided.
func (c *Connector) ListObjectMetadata( // nolint:cyclop,funlen
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	ctx = logging.With(ctx, "connector", "hubspot")

	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	// Use goroutines to fetch metadata for each object in parallel
	metadataChannel := make(chan *objectMetadataResult, len(objectNames))
	errChannel := make(chan *objectMetadataError, len(objectNames))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, objectName := range objectNames {
		go func(object string) {
			objectMetadata, err := c.describeObject(ctx, object)
			if err != nil {
				errChannel <- &objectMetadataError{
					ObjectName: object,
					Error:      err,
				}

				return
			}

			// Send object metadata to metadataChannel
			metadataChannel <- &objectMetadataResult{
				ObjectName: object,
				Response:   *objectMetadata,
			}
		}(objectName)
	}

	// Collect metadata for each object
	objectsMap := &common.ListObjectMetadataResult{}
	objectsMap.Result = make(map[string]common.ObjectMetadata)
	objectsMap.Errors = make(map[string]error)

	for range objectNames {
		select {
		// Add object metadata to objectsMap
		case objectMetadataResult := <-metadataChannel:
			objectsMap.Result[objectMetadataResult.ObjectName] = objectMetadataResult.Response
		case objectMetadataError := <-errChannel:
			objectsMap.Errors[objectMetadataError.ObjectName] = objectMetadataError.Error
		}
	}

	return objectsMap, nil
}

// describeObject returns object metadata for the given object name.
func (c *Connector) describeObject(ctx context.Context, objectName string) (*common.ObjectMetadata, error) {
	relativeURL := strings.Join([]string{"properties", objectName}, "/")

	u, err := c.getURL(relativeURL)
	if err != nil {
		return nil, err
	}

	rsp, err := c.Client.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("error fetching HubSpot fields: %w", err)
	}

	resp, err := common.UnmarshalJSON[fieldDescriptionResponse](rsp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling object metadata response into JSON: %w", err)
	}

	return common.NewObjectMetadata(
		objectName, resp.transformToFields(),
	), nil
}

func (c *Connector) GetPostAuthInfo(
	ctx context.Context,
) (*common.PostAuthInfo, error) {
	accInfo, resp, err := c.GetAccountInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching HubSpot account info: %w", err)
	}

	return &common.PostAuthInfo{
		ProviderWorkspaceRef: strconv.Itoa(accInfo.PortalId),
		RawResponse:          resp,
	}, nil
}

type AccountInfo struct {
	PortalId              int    `json:"portalId"`
	TimeZone              string `json:"timeZone"`
	CompanyCurrency       string `json:"companyCurrency"`
	AdditionalCurrencies  []string
	UTCOffset             string `json:"utcOffset"`
	UTCOffsetMilliseconds int    `json:"utcOffsetMilliseconds"`
	UIDomain              string `json:"uiDomain"`
	DataHostingLocation   string `json:"dataHostingLocation"`
}

func (c *Connector) GetAccountInfo(ctx context.Context) (*AccountInfo, *common.JSONHTTPResponse, error) {
	ctx = logging.With(ctx, "connector", "hubspot")

	resp, err := c.Client.Get(ctx, "account-info/v3/details")
	if err != nil {
		return nil, resp, fmt.Errorf("error fetching HubSpot token info: %w", err)
	}

	accountInfo, err := common.UnmarshalJSON[AccountInfo](resp)
	if err != nil {
		return nil, resp, fmt.Errorf("error unmarshalling account info response into JSON: %w", err)
	}

	return accountInfo, resp, nil
}

type fieldDescriptionResponse struct {
	Results []fieldDescription `json:"results"`
}

type fieldDescription struct {
	UpdatedAt            time.Time                 `json:"updatedAt"`
	CreatedAt            time.Time                 `json:"createdAt"`
	Name                 string                    `json:"name"`
	Label                string                    `json:"label"`
	Type                 string                    `json:"type"`
	FieldType            string                    `json:"fieldType"`
	Description          string                    `json:"description"`
	GroupName            string                    `json:"groupName"`
	Options              []fieldEnumerationOption  `json:"options"`
	DisplayOrder         int                       `json:"displayOrder"`
	Calculated           bool                      `json:"calculated"`
	ExternalOptions      bool                      `json:"externalOptions"`
	HasUniqueValue       bool                      `json:"hasUniqueValue"`
	Hidden               bool                      `json:"hidden"`
	HubspotDefined       bool                      `json:"hubspotDefined"`
	ModificationMetadata fieldModificationMetadata `json:"modificationMetadata"`
	FormField            bool                      `json:"formField"`
	DataSensitivity      string                    `json:"dataSensitivity"`
}

type fieldEnumerationOption struct {
	Label        string `json:"label"`
	Value        string `json:"value"`
	Description  string `json:"description"`
	DisplayOrder int    `json:"displayOrder"`
	Hidden       bool   `json:"hidden"`
}

type fieldModificationMetadata struct {
	Archivable         bool `json:"archivable"`
	ReadOnlyDefinition bool `json:"readOnlyDefinition"`
	ReadOnlyValue      bool `json:"readOnlyValue"`
}

func (r fieldDescriptionResponse) transformToFields() map[string]common.FieldMetadata {
	fieldsMap := make(map[string]common.FieldMetadata)

	for _, field := range r.Results {
		fieldName := strings.ToLower(field.Name)
		fieldsMap[fieldName] = field.transformToFieldMetadata()
	}

	return fieldsMap
}

// transformToFieldMetadata converts Provider model of a field into Ampersand's common.FieldMetadata.
// This normalizes provider response to the unified standard across all providers.
func (o fieldDescription) transformToFieldMetadata() common.FieldMetadata {
	var (
		valueType common.ValueType
		values    []common.FieldValue
	)

	// Based on type and field type properties from Hubspot object model map value to Ampersand value type.
	switch o.Type {
	case "string":
		valueType = common.ValueTypeString
	case "number":
		valueType = common.ValueTypeFloat
	case "bool":
		valueType = common.ValueTypeBoolean
	case "datetime":
		valueType = common.ValueTypeDateTime
	case "enumeration":
		valueType, values = o.implyEnumerationType()
		// Enumeration type means there are predefined field values.
	default:
		// ex: object_coordinates, phone_number
		valueType = common.ValueTypeOther
	}

	return common.FieldMetadata{
		DisplayName:  o.Label,
		ValueType:    valueType,
		ProviderType: o.Type + "." + o.FieldType,
		ReadOnly:     o.ModificationMetadata.ReadOnlyValue,
		Values:       values,
	}
}

func (o fieldDescription) implyEnumerationType() (common.ValueType, []common.FieldValue) {
	var values []common.FieldValue

	if len(o.Options) != 0 {
		// List of values is not nil, at least one option exists.
		values = make([]common.FieldValue, len(o.Options))
		for index, option := range o.Options {
			values[index] = common.FieldValue{
				Value:        option.Value,
				DisplayValue: option.Label,
			}
		}
	}

	switch o.FieldType {
	case "checkbox":
		return common.ValueTypeMultiSelect, values
	case "booleancheckbox":
		// Boolean values are ignored.
		return common.ValueTypeBoolean, nil
	case "radio":
		return common.ValueTypeSingleSelect, values
	case "select":
		return common.ValueTypeSingleSelect, values
	default:
		// ex: enumeration.calculation_equation
		return common.ValueTypeOther, values
	}
}
