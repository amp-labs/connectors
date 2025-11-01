package crm

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas []byte

	FileManager = scrapper.NewExtendedMetadataFileManager[staticschema.FieldMetadataMapV2, any]( // nolint:gochecknoglobals,lll
		schemas, fileconv.NewSiblingFileLocator())

	// Schemas is cached Object schemas.
	Schemas = Schema{ // nolint:gochecknoglobals
		Metadata: FileManager.MustLoadSchemas(),
	}
)

type Schema struct {
	*staticschema.Metadata[staticschema.FieldMetadataMapV2, any]
}

func (s *Schema) Select(objectNames []string) (*common.ListObjectMetadataResult, error) {
	return s.Metadata.Select(providers.ModulePipedriveCRM, objectNames)
}

var metadataDiscoveryEndpoints = datautils.Map[string, string]{ // nolint: gochecknoglobals
	"activities":    "activityFields",
	"deals":         "dealFields",
	"products":      "productFields",
	"persons":       "personFields",
	"organizations": "organizationFields",
	"notes":         "noteFields",
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

		// we only add this limit incase we're sampling fields from actual data.
		if !metadataDiscoveryEndpoints.Has(obj) {
			mdt, err := Schemas.SelectOne(providers.ModulePipedriveCRM, obj)
			if err != nil {
				objectMetadata.Errors[obj] = err

				continue
			}

			objectMetadata.Result[obj] = *mdt

			continue
		}

		res, err := a.Client.Get(ctx, url.String())
		if err != nil {
			objectMetadata.Errors[obj] = err

			continue
		}

		data, err := a.parseMetadata(res, obj)
		if err != nil {
			objectMetadata.Errors[obj] = err
		}

		if data != nil {
			objectMetadata.Result[obj] = *data
		}
	}

	return objectMetadata, nil
}

// metadataMapper constructs the metadata fields to a new map and returns it.
// Returns an error if it faces any in unmarshalling the response.
func (a *Adapter) parseMetadata( // nolint: gocognit,gocyclo,cyclop,funlen
	resp *common.JSONHTTPResponse, obj string,
) (*common.ObjectMetadata, error) {
	mdt := &common.ObjectMetadata{
		DisplayName: naming.CapitalizeFirstLetter(obj),
		FieldsMap:   make(map[string]string),
		Fields:      make(common.FieldsMetadata),
	}

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

	// For now we need to manually add & remove the v1 fields since the user is using v2 APIs.
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
