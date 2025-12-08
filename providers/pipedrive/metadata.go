package pipedrive

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers/pipedrive/metadata"
)

const (
	enum  = "enum"
	set   = "set"
	notes = "notes"
)

type metadataFields struct {
	Data []fieldResults `json:"data"`
}

type fieldResults struct {
	ID              int       `json:"id"`
	Key             string    `json:"key"`
	Name            string    `json:"name"`
	FieldType       string    `json:"field_type"`        //nolint:tagliatelle
	BulkEditAllowed bool      `json:"bulk_edit_allowed"` //nolint:tagliatelle
	Options         []options `json:"options"`
}

// options represents the set of values one can use for enum, sets data Types.
// this oly works for objects: notes, activities, organizations, deals, products, persons.
type options struct {
	ID    any    `json:"id,omitempty"` // this can be an int,bool,string
	Label string `json:"label,omitempty"`
	Color string `json:"color,omitempty"`
	AltId string `json:"alt_id,omitempty"` //nolint:tagliatelle
}

// ListObjectMetadata returns metadata for an object by sampling an object from Pipedrive's API.
// If that fails, it generates object metadata by parsing Pipedrive's OpenAPI files.
func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	objMetadata := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, obj := range objectNames {
		url, err := c.constructMetadataURL(obj)
		if err != nil {
			return nil, err
		}

		// we only add this limit incase we're sampling fields from actual data.
		if !metadataDiscoveryEndpoints.Has(obj) {
			// Limiting the response data to 1 record.
			// we only use 1 record for the metadata generation.
			// no need to query several records.
			url.WithQueryParam(limitQuery, "1")
		}

		res, err := c.Client.Get(ctx, url.String())
		if err != nil {
			objMetadata.Errors[obj] = err

			continue
		}

		data, err := parseMetadata(res, c.Module.ID, obj)
		if err != nil {
			objMetadata.Errors[obj] = err
		}

		objMetadata.Result[obj] = *data
	}

	return &objMetadata, nil
}

// metadataMapper constructs the metadata fields to a new map and returns it.
// Returns an error if it faces any in unmarshalling the response.
func parseMetadata(
	resp *common.JSONHTTPResponse, moduleID common.ModuleID, obj string,
) (*common.ObjectMetadata, error) {
	mdt := &common.ObjectMetadata{
		DisplayName: naming.CapitalizeFirstLetter(obj),
		FieldsMap:   make(map[string]string),
		Fields:      make(common.FieldsMetadata),
	}

	var err error

	if !metadataDiscoveryEndpoints.Has(obj) {
		// we currently use static schema all objects, excepts for those having
		// discovery metadata endpoints.
		mdt, err = metadata.Schemas.SelectOne(moduleID, obj)
		if err != nil {
			return nil, err
		}
	} else {
		response, err := common.UnmarshalJSON[metadataFields](resp)
		if err != nil {
			return nil, err
		}

		for _, fldRcd := range response.Data {
			mdtFlds := &common.FieldMetadata{
				DisplayName:  fldRcd.Name,
				ProviderType: fldRcd.FieldType,
				ValueType:    nativeValueType(fldRcd.FieldType),
				// All editable fields can be edited together in bulky edit dashboard.
				ReadOnly: goutils.Pointer(!fldRcd.BulkEditAllowed),
			}

			// process enums and sets fields
			processFieldOptions(mdtFlds, fldRcd, obj)

			// Add it to the objects metadatas
			mdt.AddFieldMetadata(fldRcd.Key, *mdtFlds)
		}

		// Ensure the response data array, has at least 1 record.
		// If there is no data, we use only the static schema file.
		if len(response.Data) == 0 {
			return mdt, nil
		}
	}

	return mdt, nil
}

func processFieldOptions(mdtFlds *common.FieldMetadata, fldRcd fieldResults, obj string) {
	if fldRcd.FieldType == enum || fldRcd.FieldType == set {
		for _, opt := range fldRcd.Options {
			if obj == notes && notesFlagFields.Has(fldRcd.Key) {
				mdtFlds.Values = append(mdtFlds.Values, common.FieldValue{
					Value:        opt.Label,
					DisplayValue: opt.Label,
				})

				continue
			}

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
