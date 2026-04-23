package apollo

import (
	"context"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// Write creates/updates records in apolllo.
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) { //nolint: cyclop,lll
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	var write common.WriteMethod

	url, err := c.getAPIURL(config.ObjectName, writeOp)
	if err != nil {
		return nil, err
	}
	// sets post as default
	write = c.Client.Post

	// Add customFields mappings.
	recordData, okay := config.RecordData.(map[string]any)
	if !okay {
		return nil, fmt.Errorf("err: unexpected data type %w", common.ErrMissingExpectedValues)
	}

	// we check for custom fields in any scenario when we have 0 custom fields for a particular object.
	if usesFieldsResource.Has(config.ObjectName) && c.customFields[config.ObjectName] == nil { // nolint: nestif
		// If we're reading this for the first time, we make a call to retrieve
		// custom fields, add them and their labels in the connector instance field customFields.
		if err := c.retrieveCustomFields(ctx, config.ObjectName); err != nil {
			return nil, err
		}

		flds := c.customFields[config.ObjectName]

		if flds != nil {
			cstFlds := make(map[string]any)

			for _, label := range flds {
				if val, exists := recordData[label.fld]; exists {
					cstFlds[label.customMachineField] = val
				}
			}

			if len(cstFlds) > 0 {
				recordData["typed_custom_fields"] = cstFlds
			}
		}
	}

	// prepares the updating data request.
	if len(config.RecordId) > 0 {
		url = url.AddPath(config.RecordId)

		write = c.Client.Patch
	}

	json, err := write(ctx, url.String(), recordData)
	if err != nil {
		return nil, err
	}

	body, ok := json.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	return c.constructWriteResult(body, config.ObjectName)
}

func (c *Connector) constructWriteResult(body *ajson.Node, objName string) (*common.WriteResult, error) {
	objName = constructSupportedObjectName(objName)

	// API Response contains a json object having a singular objectName key with the
	// created/updated details in it.
	obj := naming.NewSingularString(objName)

	respObject, err := jsonquery.New(body).ObjectRequired(obj.String())
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(respObject).StrWithDefault("id", "")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(respObject)
	if err != nil {
		return nil, err
	}

	if customFields, exists := data["typed_custom_fields"]; exists {
		flds, ok := customFields.(map[string]any)
		if ok {
			for fld, val := range flds {
				for _, cstFld := range c.customFields[objName] {
					if fld == cstFld.customMachineField {
						data[cstFld.fld] = val
					}
				}
			}
		}
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}

// retrieveCustomFields fetches the ObjectMetadata using the fields API
// https://docs.apollo.io/reference/get-a-list-of-fields this requires master API key.
func (c *Connector) retrieveCustomFields(ctx context.Context, objectName string, //nolint: cyclop
) error {
	var response *FieldsResponse

	url, err := c.getAPIURL(fields, readOp)
	if err != nil {
		return err
	}

	url.WithQueryParam("source", custom)

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return err
	}

	response, err = common.UnmarshalJSON[FieldsResponse](resp)
	if err != nil {
		return err
	}

	for _, fld := range response.Fields {
		c.customFields.Set(objectName, fld.Label, fld.Label, strings.TrimPrefix(fld.Id, fld.Modality+"."))
	}

	return nil
}
