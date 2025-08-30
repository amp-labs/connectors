package bitbucket

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	restAPIVersion   = "2.0"
	perPageQuery     = "pagelen"
	metadataPageSize = "1"
)

type httpResponse struct {
	PageLen int              `json:"pagelen"`
	Page    int              `json:"page"`
	Size    int              `json:"size"`
	Values  []map[string]any `json:"values"`
}

func (c *Connector) buildSingleHandlerRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, objectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(perPageQuery, metadataPageSize)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseSingleHandlerResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		Fields:      make(common.FieldsMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	data, err := common.UnmarshalJSON[httpResponse](response)
	if err != nil {
		return nil, err
	}

	if len(data.Values) < 1 {
		return nil, common.ErrMissingFields
	}

	for fld := range data.Values[0] {
		objectMetadata.FieldsMap[fld] = fld
	}

	return &objectMetadata, nil
}
