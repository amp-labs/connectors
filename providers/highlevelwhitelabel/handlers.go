package highlevelwhitelabel

import (
	"context"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, objectName)
	if err != nil {
		return nil, err
	}

	if objectsWithLocationIdInParam.Has(objectName) {
		url.WithQueryParam("locationId", c.locationId)
	}

	if objectWithAltTypeAndIdQueryParam.Has(objectName) {
		url.WithQueryParam("altId", c.locationId)
		url.WithQueryParam("altType", "location")
	}

	if paginationObjects.Has(objectName) {
		url.WithQueryParam("limit", "1")

		if objectWithSkipQueryParam.Has(objectName) {
			url.WithQueryParam("skip", "0")
		} else {
			url.WithQueryParam("offset", "0")
		}
	}

	// For single-segment paths (e.g., "businesses"), the URL must have a trailing slash at the end.
	// Example: https://highlevel.stoplight.io/docs/integrations/a8db8afcbe0a3-get-businesses-by-location
	//
	// For multi-segment paths (e.g., "calendars/groups"), the URL does not require a trailing slash.
	// Example: https://highlevel.stoplight.io/docs/integrations/89e47b6c05e67-get-groups
	if !(strings.Contains(objectName, "/")) {
		urlRaw, err := url.ToURL()
		if err != nil {
			return nil, err
		}

		urlRaw.Path = urlRaw.Path + "/" // nolint:gocritic

		url, err = urlbuilder.FromRawURL(urlRaw)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Version", apiVersion)

	return req, nil
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

	nodePath := objectsNodePath.Get(objectName)

	objectResponse, err := jsonquery.New(node).ArrayRequired(nodePath)
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

	for field := range data[0] {
		objectMetadata.FieldsMap[field] = field
	}

	return &objectMetadata, nil
}
