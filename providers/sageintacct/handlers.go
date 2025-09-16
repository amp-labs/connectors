package sageintacct

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	apiPrefix  = "ia/api"
	apiVersion = "v1"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiPrefix, apiVersion, "services/core/model")
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("name", objectName)
	url.WithQueryParam("version", apiVersion)

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetter(objectName),
	}

	bodyNode, ok := response.Body()
	if !ok {
		return nil, common.ErrFailedToUnmarshalBody
	}

	resultNode, err := bodyNode.GetKey("ia::result")
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	// api returns array when object is not supported.
	if resultNode.IsArray() {
		return nil, common.ErrObjectNotSupported
	}

	res, err := common.UnmarshalJSON[SageIntacctMetadataResponse](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	for fieldName, fieldDef := range res.Result.Fields {
		objectMetadata.Fields[fieldName] = common.FieldMetadata{
			DisplayName:  naming.CapitalizeFirstLetterEveryWord(fieldName),
			ValueType:    mapSageIntacctTypeToValueType(fieldDef.Type),
			ProviderType: fieldDef.Type,
			ReadOnly:     fieldDef.ReadOnly,
			Values:       mapValuesFromEnum(fieldDef),
		}
	}

	if len(res.Result.Groups) > 0 {
		for groupName := range res.Result.Groups {
			objectMetadata.Fields[groupName] = common.FieldMetadata{
				DisplayName:  naming.CapitalizeFirstLetterEveryWord(groupName),
				ValueType:    common.ValueTypeOther,
				ProviderType: "object",
				ReadOnly:     false,
				Values:       []common.FieldValue{},
			}
		}
	}

	return &objectMetadata, nil
}
