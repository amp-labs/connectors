package servicenow

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
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

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, params.ObjectName)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(response,
		common.ExtractRecordsFromPath("result"),
		getNextRecordsURL(response.Headers.Get("Link")),
		common.GetMarshaledData,
		params.Fields,
	)
}

func getNextRecordsURL(linkHeader string) common.NextPageFunc {
	return func(n *ajson.Node) (string, error) {
		return ParseNexPageLinkHeader(linkHeader)
	}
}

// ParseNexPageLinkHeader extracts the next page URL from the Link Header response.
func ParseNexPageLinkHeader(linkHeader string) (string, error) {
	if linkHeader == "" {
		return "", nil // this indicates we're done.
	}

	links := strings.Split(linkHeader, ",")
	// [<https://dev269415.service-now.com/api/now/v2/table/incident?sysparm_limit=1&sysparm_offset=0>;rel="next" ...]
	for _, link := range links {
		if strings.Contains(link, `rel="next"`) {
			parts := strings.Split(link, ";")
			rawURL := strings.TrimSpace(parts[0])
			rawURL = strings.TrimPrefix(rawURL, "<")
			rawURL = strings.TrimSuffix(rawURL, ">")

			// Parse the URL to ensure it's valid
			parsedURL, err := url.Parse(rawURL)
			if err != nil {
				return "", fmt.Errorf("failed to parse URL: %w", err)
			}

			return parsedURL.String(), nil
		}
	}

	return "", nil
}
