package snapchatads

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := c.constructURL(objectName)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	node, ok := response.Body() // nolint:varnamelen
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	objectResponse, err := jsonquery.New(node).ArrayRequired(objectName)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ArrayToMap(objectResponse)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	objKey := naming.NewSingularString(objectName).String()

	// Extract and assert the inner map
	innerData, ok := data[0][objKey].(map[string]any)
	if !ok {
		return nil, ErrObjNotFound
	}

	for field := range innerData {
		// TODO fix deprecated
		objectMetadata.FieldsMap[field] = field // nolint:staticcheck
	}

	return &objectMetadata, nil
}
