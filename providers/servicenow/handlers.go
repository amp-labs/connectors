package servicenow

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

type responseData struct {
	Result []map[string]any `json:"result"`
	// Other fields
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, objectName)
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

	res, err := common.UnmarshalJSON[responseData](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(res.Result) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	// Using the first result data to generate the metadata.
	for k := range res.Result[0] {
		objectMetadata.FieldsMap[k] = k
	}

	return &objectMetadata, nil
}
