package zohocrm

import (
	"context"
	"fmt"
	"sync"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

// restMetadataEndpoint is the resource for retrieving metadata details.
// doc: https://www.zoho.com/crm/developer/docs/api/v6/field-meta.html
const restMetadataEndpoint = "settings/fields"

type metadataFields struct {
	Fields []map[string]any `json:"fields"`
}

type metadataFieldsV2 struct {
	Fields []field `json:"fields"`
}

//nolint:tagliatelle
type field struct {
	Name           string        `json:"api_name"`
	DisplayName    string        `json:"field_label"`
	Type           string        `json:"data_type"`
	ReadOnly       bool          `json:"read_only"`
	PickListValues []fieldValues `json:"pick_list_values,omitempty"`
	// The rest metadata details
}

//nolint:tagliatelle
type fieldValues struct {
	DisplayValue string `json:"display_value"`
	ActualValue  string `json:"actual_value"`
}

func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	var (
		wg sync.WaitGroup //nolint: varnamelen
		mu sync.Mutex     //nolint: varnamelen
	)

	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	objectMetadata := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata, len(objectNames)),
		Errors: make(map[string]error, len(objectNames)),
	}

	wg.Add(len(objectNames))

	for _, object := range objectNames {
		go func(object string) {
			metadata, err := c.getMetadata(ctx, object)
			if err != nil {
				mu.Lock()
				objectMetadata.Errors[object] = err
				mu.Unlock()
				wg.Done()

				return
			}

			mu.Lock()
			objectMetadata.Result[object] = *metadata
			mu.Unlock()

			wg.Done()
		}(object)
	}

	// Wait for all goroutines to finish their calls.
	wg.Wait()

	return &objectMetadata, nil
}

func (c *Connector) fetchFieldMetadata(ctx context.Context, capObj string) (*common.JSONHTTPResponse, error) {
	url, err := c.getAPIURL(restMetadataEndpoint)
	if err != nil {
		return nil, err
	}

	// setting this, returns both used and unused fields
	url.WithQueryParam("type", "all")
	url.WithQueryParam("module", capObj)

	return c.Client.Get(ctx, url.String())
}

func (c *Connector) getMetadata(ctx context.Context, objectName string) (*common.ObjectMetadata, error) {
	capObj := naming.CapitalizeFirstLetterEveryWord(objectName)

	resp, err := c.fetchFieldMetadata(ctx, capObj)
	if err != nil {
		return nil, fmt.Errorf("error fetching metadata: %w", err)
	}

	metadata, err := parseMetadataResponse(resp, capObj)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func parseMetadataResponse(resp *common.JSONHTTPResponse, objectName string) (*common.ObjectMetadata, error) {
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
			ValueType:    nativeType(fld.Type),
			ProviderType: fld.Type,
			ReadOnly:     fld.ReadOnly,
			Values:       fieldValues,
		}

		metadata.AddFieldMetadata(fld.Name, mdt)
	}

	return metadata, nil
}

func nativeType(typ string) common.ValueType {
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
