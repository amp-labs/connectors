package pipedrive

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/pipedrive/metadata"
)

const (
	enum          = "enum"
	set           = "set"
	notes         = "notes"
	activities    = "activities"
	deals         = "deals"
	products      = "products"
	organizations = "organizations"
	persons       = "persons"
	pipelines     = "pipelines"
	stages        = "stages"
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
	AltId string `json:"alt_id,omitempty"`
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

		data, err := c.parseMetadata(res, c.moduleID, obj)
		if err != nil {
			objMetadata.Errors[obj] = err
		}

		objMetadata.Result[obj] = *data
	}

	return &objMetadata, nil
}

// metadataMapper constructs the metadata fields to a new map and returns it.
// Returns an error if it faces any in unmarshalling the response.
func (c *Connector) parseMetadata( // nolint: gocognit,gocyclo,cyclop,funlen
	resp *common.JSONHTTPResponse, moduleID common.ModuleID, obj string,
) (*common.ObjectMetadata, error) {
	mdt := &common.ObjectMetadata{
		DisplayName: naming.CapitalizeFirstLetter(obj),
		FieldsMap:   make(map[string]string),
		Fields:      make(common.FieldsMetadata),
	}

	var err error

	if !metadataDiscoveryEndpoints.Has(obj) { //nolint: nestif
		if c.moduleID == providers.PipedriveV2 {
			// we currently use static schema all objects, excepts for those having
			// discovery metadata endpoints.
			mdt, err = metadata.SchemasV2.SelectOne(moduleID, obj)
			if err != nil {
				return nil, err
			}
		} else {
			// we currently use static schema all objects, excepts for those having
			// discovery metadata endpoints.
			mdt, err = metadata.Schemas.SelectOne(moduleID, obj)
			if err != nil {
				return nil, err
			}
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
				ReadOnly: !fldRcd.BulkEditAllowed,
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

		if c.moduleID == providers.PipedriveV2 {
			// For now we need to manually add & remove the v1 fields incase the user is using v2 APIs.
			switch obj {
			case activities:
				for _, fld := range activityRemovedFields.List() {
					mdt.RemoveFieldMetadata(fld)
				}

				for prvFld, newFld := range activityRenamedFields {
					mdt.Fields[newFld] = mdt.Fields[prvFld]
					mdt.RemoveFieldMetadata(prvFld)
				}

			case deals:
				for _, fld := range dealRemovedFields.List() {
					mdt.RemoveFieldMetadata(fld)
				}

				for prvFld, newFld := range dealRenamedFields {
					mdt.Fields[newFld] = mdt.Fields[prvFld]
					mdt.RemoveFieldMetadata(prvFld)
				}

				for fld, typ := range dealAddedFields {
					mdt.AddFieldMetadata(fld, common.FieldMetadata{
						DisplayName: fld,
						ValueType:   typ,
					})
				}

			case persons:
				for _, fld := range personRemovedFields.List() {
					mdt.RemoveFieldMetadata(fld)
				}

				for prvFld, newFld := range personRenamedFields {
					mdt.Fields[newFld] = mdt.Fields[prvFld]
					mdt.RemoveFieldMetadata(prvFld)
				}

				for fld, typ := range personAddedFields {
					mdt.AddFieldMetadata(fld, common.FieldMetadata{
						DisplayName: fld,
						ValueType:   typ,
					})
				}
			case stages:
				for _, fld := range stageRemovedFields.List() {
					mdt.RemoveFieldMetadata(fld)
				}

				for prvFld, newFld := range stageRenamedFields {
					mdt.Fields[newFld] = mdt.Fields[prvFld]
					mdt.RemoveFieldMetadata(prvFld)
				}
			case pipelines:
				for _, fld := range pipelineRemovedFields.List() {
					mdt.RemoveFieldMetadata(fld)
				}

				for prvFld, newFld := range pipelineRenamedFields {
					mdt.Fields[newFld] = mdt.Fields[prvFld]
					mdt.RemoveFieldMetadata(prvFld)
				}
			case organizations:
				for _, fld := range organizationRemovedFields.List() {
					mdt.RemoveFieldMetadata(fld)
				}

				for prvFld, newFld := range organizationRenamedFields {
					mdt.Fields[newFld] = mdt.Fields[prvFld]
					mdt.RemoveFieldMetadata(prvFld)
				}

				for fld, typ := range organizationAddedFields {
					mdt.AddFieldMetadata(fld, common.FieldMetadata{
						DisplayName: fld,
						ValueType:   typ,
					})
				}

			case products:
				for _, fld := range productRemovedFields.List() {
					mdt.RemoveFieldMetadata(fld)
				}

				for prvFld, newFld := range productRenamedFields {
					mdt.Fields[newFld] = mdt.Fields[prvFld]
					mdt.RemoveFieldMetadata(prvFld)
				}

				for fld, typ := range productAddedFields {
					mdt.AddFieldMetadata(fld, common.FieldMetadata{
						DisplayName: fld,
						ValueType:   typ,
					})
				}
			}
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
