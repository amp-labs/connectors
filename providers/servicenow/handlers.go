package servicenow

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// 0. Support Table API, Internet shows this is the most used API.
// 1. Support the whole or parts of the `now` namespace.
// 2. Look into adding more namespace (modules) supports.

type responseData struct {
	Result []map[string]any `json:"result"`
	// Other fields
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	module := supportedModules[c.Module()]

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, module.Path(), objectName)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
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
