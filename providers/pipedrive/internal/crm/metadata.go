package crm

import (
	"context"
	_ "embed"
	"fmt"
	"reflect"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
)

/*
Currently supported v2 objects:
 - Activities
 - Deals
 - ItemSearch
 - Organizations
 - Persons
 - Pipelines
 - Products
 - Stages
*/

var metadataDiscoveryEndpoints = datautils.Map[string, string]{ // nolint: gochecknoglobals
	"activities":    "activityFields",
	"deals":         "dealFields",
	"products":      "productFields",
	"persons":       "personFields",
	"organizations": "organizationFields",
	// leadFields, NoteFields still uses v1fields.
}

func (a *Adapter) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	objectMetadata := &common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, obj := range objectNames {
		url, err := a.constructMetadataURL(obj)
		if err != nil {
			return nil, err
		}

		response, err := a.Client.Get(ctx, url.String())
		if err != nil {
			objectMetadata.Errors[obj] = err

			continue
		}

		metadata, err := a.parseMetadata(response, obj)
		if err != nil {
			objectMetadata.Errors[obj] = err
		}

		objectMetadata.Result[obj] = metadata
	}

	return objectMetadata, nil
}

// metadataMapper constructs the metadata fields to a new map and returns it.
// Returns an error if it faces any in unmarshalling the response.
func (a *Adapter) parseMetadata( // nolint: gocognit,gocyclo,cyclop,funlen
	resp *common.JSONHTTPResponse, obj string,
) (common.ObjectMetadata, error) {
	mdt := common.ObjectMetadata{
		DisplayName: naming.CapitalizeFirstLetter(obj),
		FieldsMap:   make(map[string]string),
		Fields:      make(common.FieldsMetadata),
	}

	if !metadataDiscoveryEndpoints.Has(obj) {
		response, err := common.UnmarshalJSON[records](resp)
		if err != nil {
			return common.ObjectMetadata{}, err
		}

		if len(response.Data) == 0 {
			return common.ObjectMetadata{}, common.ErrMissingExpectedValues
		}

		firstRecord := response.Data[0]
		for fld, val := range firstRecord {
			mdt.Fields[fld] = common.FieldMetadata{
				DisplayName: fld,
				ValueType:   inferValue(val),
			}
		}

		return mdt, nil
	}

	response, err := common.UnmarshalJSON[metadataFields](resp)
	if err != nil {
		return common.ObjectMetadata{}, err
	}

	for _, fldRcd := range response.Data {
		req := !fldRcd.IsOptional

		mdtFlds := &common.FieldMetadata{
			DisplayName:  fldRcd.Name,
			IsCustom:     &fldRcd.IsCustom,
			IsRequired:   &(req),
			ProviderType: fldRcd.FieldType,
			ValueType:    nativeValueType(fldRcd.FieldType),
		}

		// process enums and sets fields
		processFieldOptions(mdtFlds, fldRcd)

		// Add it to the objects metadatas
		mdt.AddFieldMetadata(fldRcd.Code, *mdtFlds)
	}

	// Ensure the response data array, has at least 1 record.
	// If there is no data, we return.
	if len(response.Data) == 0 {
		return mdt, nil
	}

	return mdt, nil
}

func processFieldOptions(mdtFlds *common.FieldMetadata, fldRcd fieldResults) {
	if fldRcd.FieldType == enum || fldRcd.FieldType == set {
		for _, opt := range fldRcd.Options {
			mdtFlds.Values = append(mdtFlds.Values, common.FieldValue{
				Value:        fmt.Sprint(opt.ID),
				DisplayValue: opt.Label,
			})
		}
	}
}

func nativeValueType(providerTyp string) common.ValueType {
	switch providerTyp {
	case "int":
		return common.ValueTypeInt
	case "varchar", "text":
		return common.ValueTypeString
	case "enum":
		return common.ValueTypeSingleSelect
	case "set":
		return common.ValueTypeMultiSelect
	case "double":
		return common.ValueTypeFloat
	case "date":
		return common.ValueTypeDate
	case "time":
		return common.ValueTypeDateTime
	default:
		return common.ValueTypeOther
	}
}

func inferValue(value any) common.ValueType {
	v := reflect.ValueOf(value)

	switch v.Kind() { //nolint: exhaustive
	case reflect.String:
		return common.ValueTypeString
	case reflect.Float64:
		return common.ValueTypeFloat
	case reflect.Bool:
		return common.ValueTypeBoolean
	case reflect.Slice:
		return common.ValueTypeOther
	case reflect.Map:
		return common.ValueTypeOther
	default:
		return common.ValueTypeOther
	}
}
