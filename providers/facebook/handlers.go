package facebook

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const apiVersion = "v19.0"

type ResponseData struct {
	Data []map[string]any `json:"data"`
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	urlPath := c.constructURL(objectName)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, urlPath)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	// Adding accept because this connector sending response text/javascript.
	request.Header.Add("Accept", "*/*")

	return request, nil
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

	resp, err := common.UnmarshalJSON[ResponseData](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	for field := range resp.Data[0] {
		objectMetadata.FieldsMap[field] = field
	}

	return &objectMetadata, nil
}
