package apollo

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

var (
	perPage          = "per_page" //nolint:gochecknoglobals
	metadataPageSize = "1"        //nolint:gochecknoglobals
	fields           = "fields"   //nolint:gochecknoglobals
)

type FieldsResponse struct {
	Fields []Field `json:"fields"`
}

type Field struct {
	Id        string `json:"id"`
	Category  string `json:"category"`
	Editable  bool   `json:"editable"`
	Example   any    `json:"example"`
	FieldName string `json:"field_name"`
	Source    string `json:"source"`
	Modality  string `json:"modality"`
	Type      string `json:"type"`
	Label     string `json:"label"`
	// Other fields
}

// ListObjectMetadata creates metadata of object via reading objects using Apollo API.
func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	// Ensure that objectNames is not empty
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	metadataResult, err := c.requestMetadata(ctx, objectNames)
	if err != nil {
		return nil, err
	}

	return metadataResult, nil
}

func (c *Connector) requestMetadata(ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	metadataResult := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, objectName := range objectNames {
		// for objects: Accounts, Contacts we use the fields endpoint to construct the metadatas
		// We have to make 3 API calls for the standard(system), custom, crm fields.
		if usesFieldsResource.Has(objectName) {
			metadata, err := c.retrieveFields(ctx)
			if err != nil {
				metadataResult.Errors[objectName] = err
			}

			metadata.DisplayName = objectName
			metadataResult.Result[objectName] = *metadata

			continue
		}

		url, err := c.getAPIURL(objectName, readOp)
		if err != nil {
			return nil, err
		}

		// Limiting the response, so as we don't have to return 100 records of data
		// when we just need 1.
		url.WithQueryParam(perPage, metadataPageSize)

		resp, err := c.Client.Get(ctx, url.String())
		if err != nil {
			metadataResult.Errors[objectName] = err

			continue
		}

		// Check nil response body, to avoid panic.
		body, ok := resp.Body()
		if !ok {
			metadataResult.Errors[objectName] = common.ErrEmptyJSONHTTPResponse

			continue
		}

		metadata, err := parseMetadataFromResponse(body, objectName)
		if err != nil {
			return nil, err
		}

		metadata.DisplayName = objectName
		metadataResult.Result[objectName] = *metadata
	}

	return &metadataResult, nil
}

func (c *Connector) retrieveFields(ctx context.Context) (*common.ObjectMetadata, error) {
	var response *FieldsResponse

	objectMetadata := common.ObjectMetadata{
		Fields: make(common.FieldsMetadata),
	}

	for _, v := range []string{"custom", "system"} {
		url, err := c.getAPIURL(fields, readOp)
		if err != nil {
			return nil, err
		}

		url.WithQueryParam("source", v)

		resp, err := c.Client.Get(ctx, url.String())
		if err != nil {
			return nil, err
		}

		response, err = common.UnmarshalJSON[FieldsResponse](resp)
		if err != nil {
			return nil, err
		}

		for _, fld := range response.Fields {
			var (
				isCustom   bool
				isEditable bool
			)

			if fld.Modality != "contact" {
				continue
			}

			if fld.Source == "custom" {
				isCustom = true
			}

			if fld.Editable {
				isEditable = true
			}

			objectMetadata.Fields[strings.TrimPrefix(fld.Id, "contact.")] = common.FieldMetadata{
				DisplayName:  fld.Label,
				ReadOnly:     &isEditable,
				ProviderType: fld.Type,
				IsCustom:     &isCustom,
				ValueType:    common.InferValueTypeFromData(fld.Example),
			}
		}
	}

	return &objectMetadata, nil
}

func parseMetadataFromResponse(body *ajson.Node, objectName string) (*common.ObjectMetadata, error) {
	objectName = constructSupportedObjectName(objectName)

	arr, err := jsonquery.New(body).ArrayOptional(objectName)
	if err != nil {
		return nil, err
	}

	fieldsMap := make(map[string]string)

	if len(arr) != 0 {
		objectResponse := arr[0].MustObject()

		// Using the result data to generate the metadata.
		for k := range objectResponse {
			fieldsMap[k] = k
		}
	}

	return &common.ObjectMetadata{
		FieldsMap: fieldsMap,
	}, nil
}
