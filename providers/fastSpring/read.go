package fastspring

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/fastspring/metadata"
	"github.com/spyzhov/ajson"
)

const (
	defaultPageSize = "50"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	if params.ObjectName == "" {
		return nil, common.ErrMissingObjects
	}

	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	// FastSpring list endpoints support basic cursor-like pagination via limit + page.
	url.WithQueryParam("limit", readhelper.PageSizeWithDefaultStr(params, defaultPageSize))

	// When no explicit page is provided via NextPage, we start from page=1.
	url.WithQueryParam("page", "1")

	return url, nil
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	recordsKey := metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), params.ObjectName)

	records := common.ExtractRecordsFromPath(recordsKey)

	return common.ParseResult(
		resp,
		records,
		nextPageFromIntegerCounter(request.URL),
		common.GetMarshaledData,
		params.Fields,
	)
}

// nextPageFromIntegerCounter builds a NextPageFunc that reads a numeric "nextPage"
// field from the response root and maps it to the "page" query parameter.
func nextPageFromIntegerCounter(previousRequestURL *url.URL) common.NextPageFunc {
	return func(root *ajson.Node) (string, error) {
		if previousRequestURL == nil {
			return "", nil
		}

		nextPage, err := jsonquery.New(root).IntegerWithDefault("nextPage", 0)
		if err != nil || nextPage == 0 {
			return "", err
		}

		cloned := *previousRequestURL
		q := cloned.Query()
		q.Set("page", strconv.FormatInt(nextPage, 10))
		cloned.RawQuery = q.Encode()

		return cloned.String(), nil
	}
}
