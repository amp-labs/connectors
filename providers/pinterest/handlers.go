package pinterest

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

type responseObject struct {
	Items    []map[string]any `json:"items"`
	Bookmark string           `json:"bookmark"`
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	urlPath := matchObjectNameToEndpointPath(objectName)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, urlPath)
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

	data, err := common.UnmarshalJSON[responseObject](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	for field := range data.Items[0] {
		objectMetadata.FieldsMap[field] = field
	}

	return &objectMetadata, nil
}

func matchObjectNameToEndpointPath(objectName string) (urlPath string) {
	switch objectName {
	// https://developers.pinterest.com/docs/api/v5/get-business_employers.
	case "employers":
		return "businesses/employers"
	// https://developers.pinterest.com/docs/api/v5/feeds-list.
	case "feeds":
		return "catalogs/feeds"
	// https://developers.pinterest.com/docs/api/v5/catalogs_product_groups-list.
	case "product_groups":
		return "catalogs/product_groups"
	// https://developers.pinterest.com/docs/api/v5/reports-stats.
	case "stats":
		return "catalogs/reports/stats"
	default:
		return objectName
	}
}
