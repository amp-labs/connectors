package zoho

import (
	"context"
	"fmt"
	"sync"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/amp-labs/connectors/providers"
)

//nolint:unused // used in ListObjectMetadata
type metadataFetcher func(ctx context.Context, objectName string) (*common.ObjectMetadata, error)

// =============================================================================
// ZohoCRM Metadata Types

// restMetadataEndpoint is the resource for retrieving metadata details.
// doc: https://www.zoho.com/crm/developer/docs/api/v6/field-meta.html
const (
	restMetadataEndpoint = "settings/fields"
	users                = "users"
	org                  = "org"
)

type metadataFields struct {
	Fields []map[string]any `json:"fields"`
}

type metadataFieldsV2 struct {
	Fields []field `json:"fields"`
}

// Response from: https://www.zoho.com/crm/developer/docs/api/v6/field-meta.html
//
//nolint:tagliatelle
type field struct {
	Name        string `json:"api_name"`
	DisplayName string `json:"field_label"`
	Type        string `json:"data_type"`
	// Whether field is read only for the current user
	// We ignore this.
	ReadOnly bool `json:"read_only"`
	// Whether field is always read only for everyone
	// This is the field we return in FieldMetadata.ReadOnly
	FieldReadOnly  bool          `json:"field_read_only"`
	PickListValues []fieldValues `json:"pick_list_values,omitempty"`
	// The rest metadata details
}

//nolint:tagliatelle
type fieldValues struct {
	DisplayValue string `json:"display_value"`
	ActualValue  string `json:"actual_value"`
}

// ==============================================================================
// ZohoDesk Metadata Types

const deskMetadataEndpoint = "organizationFields"

type deskField struct {
	Name          string       `json:"apiName"`
	DisplayLabel  string       `json:"displayLabel"`
	Type          string       `json:"type"`
	AllowedValues []deskValues `json:"allowedValues,omitempty"`
	// The rest metadata details
}

type deskMetadataFields struct {
	Data []deskField `json:"data"`
}

type deskValues struct {
	Value string `json:"value,omitempty"`
}

// ==============================================================================

func (c *Connector) ListObjectMetadata( // nolint:wsl_v5
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	var (
		mu sync.Mutex                      //nolint: varnamelen
		mf metadataFetcher = c.crmMetadata //nolint: varnamelen
	)

	if c.moduleID == providers.ModuleZohoDesk {
		mf = c.deskMetadata
	}

	if c.isServiceDeskPlusModule() {
		return c.servicedeskplusAdapter.ListObjectMetadata(ctx, objectNames)
	}

	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	objectMetadata := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata, len(objectNames)),
		Errors: make(map[string]error, len(objectNames)),
	}

	callbacks := make([]simultaneously.Job, 0, len(objectNames))

	for _, object := range objectNames {
		obj := object // capture loop variable

		callbacks = append(callbacks, func(ctx context.Context) error {
			metadata, err := mf(ctx, obj)
			if err != nil {
				mu.Lock()
				objectMetadata.Errors[obj] = err // nolint:wsl_v5
				mu.Unlock()

				return nil //nolint:nilerr // intentionally collecting errors in map, not failing fast
			}

			mu.Lock()
			objectMetadata.Result[object] = *metadata // nolint:wsl_v5
			mu.Unlock()

			return nil
		})
	}

	if err := simultaneously.DoCtx(ctx, -1, callbacks...); err != nil {
		return nil, err
	}

	return &objectMetadata, nil
}

// =============================================================================
// ZohoCRM fetching metadata fields and related details functions

func (c *Connector) fetchCRMFieldResponse(ctx context.Context, capObj string) (*common.JSONHTTPResponse, error) {
	url, err := c.getAPIURL(crmAPIVersion, restMetadataEndpoint)
	if err != nil {
		return nil, err
	}

	// setting this, returns both used and unused fields
	url.WithQueryParam("type", "all")
	url.WithQueryParam("module", capObj)

	return c.Client.Get(ctx, url.String())
}

func (c *Connector) crmMetadata(ctx context.Context, objectName string) (*common.ObjectMetadata, error) {
	capObj := objectName
	if objectName != users {
		capObj = naming.CapitalizeFirstLetterEveryWord(objectName)
	}

	resp, err := c.fetchCRMFieldResponse(ctx, capObj)
	if err != nil {
		return nil, fmt.Errorf("error fetching metadata: %w", err)
	}

	metadata, err := parseCRMMetadataResponse(resp, capObj)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func parseCRMMetadataResponse(resp *common.JSONHTTPResponse, objectName string) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[metadataFieldsV2](resp)
	if err != nil {
		return nil, err
	}

	metadata := &common.ObjectMetadata{
		DisplayName: objectName,
		Fields:      make(common.FieldsMetadata),
		FieldsMap:   make(map[string]string),
	}

	// Ranging on the fields Slice, to construct the metadata fields.
	for _, fld := range response.Fields {
		var fieldValues []common.FieldValue

		for _, opt := range fld.PickListValues {
			fieldValues = append(fieldValues, common.FieldValue{
				Value:        opt.ActualValue,
				DisplayValue: opt.DisplayValue,
			})
		}

		mdt := common.FieldMetadata{
			DisplayName:  fld.DisplayName,
			ValueType:    nativeCRMType(fld.Type),
			ProviderType: fld.Type,
			ReadOnly:     goutils.Pointer(fld.FieldReadOnly),
			Values:       fieldValues,
		}

		metadata.AddFieldMetadata(fld.Name, mdt)
	}

	return metadata, nil
}

func nativeCRMType(typ string) common.ValueType {
	switch typ {
	case "text", "textarea", "email", "phone", "website":
		return common.ValueTypeString
	case "date":
		return common.ValueTypeDate
	case "datetime":
		return common.ValueTypeDateTime
	case "boolean":
		return common.ValueTypeBoolean
	case "integer":
		return common.ValueTypeInt
	case "picklist":
		return common.ValueTypeSingleSelect
	case "multiselectpicklist":
		return common.ValueTypeMultiSelect
	default:
		return common.ValueTypeOther
	}
}

// =============================================================================
// Zoho Desk fetching metadata and related details functions.

func (c *Connector) fetchDeskFieldsResponse(ctx context.Context, objectName string) (*common.JSONHTTPResponse, error) {
	url, err := c.getAPIURL(deskAPIVersion, deskMetadataEndpoint)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("module", objectName)

	return c.Client.Get(ctx, url.String())
}

func (c *Connector) deskMetadata(ctx context.Context, objectName string) (*common.ObjectMetadata, error) {
	resp, err := c.fetchDeskFieldsResponse(ctx, objectName)
	if err != nil {
		return nil, fmt.Errorf("error fetching metadata: %w", err)
	}

	metadata, err := parseDeskMetadataResponse(resp, objectName)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func parseDeskMetadataResponse(resp *common.JSONHTTPResponse, objectName string) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[deskMetadataFields](resp)
	if err != nil {
		return nil, err
	}

	metadata := &common.ObjectMetadata{
		DisplayName: objectName,
		Fields:      make(common.FieldsMetadata),
		FieldsMap:   make(map[string]string),
	}

	// Ranging on the fields Slice, to construct the metadata fields.
	for _, fld := range response.Data {
		var fieldValues []common.FieldValue

		for _, opt := range fld.AllowedValues {
			fieldValues = append(fieldValues, common.FieldValue{
				Value:        opt.Value,
				DisplayValue: opt.Value,
			})
		}

		mdt := common.FieldMetadata{
			DisplayName:  fld.DisplayLabel,
			ValueType:    nativeDeskType(fld.Type),
			ProviderType: fld.Type,
			Values:       fieldValues,
		}

		metadata.AddFieldMetadata(fld.Name, mdt)
	}

	return metadata, nil
}

func nativeDeskType(typ string) common.ValueType {
	switch typ {
	case "Text", "Email", "Phone", "Textarea", "URL", "LargeText":
		return common.ValueTypeString
	case "Date":
		return common.ValueTypeDate
	case "DateTime":
		return common.ValueTypeDateTime
	case "Boolean":
		return common.ValueTypeBoolean
	case "Number":
		return common.ValueTypeInt
	case "Picklist":
		return common.ValueTypeSingleSelect
	case "Multiselect":
		return common.ValueTypeMultiSelect
	default:
		return common.ValueTypeOther
	}
}
