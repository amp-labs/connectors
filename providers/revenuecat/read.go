package revenuecat

import (
	"context"
	"net/http"
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/revenuecat/metadata"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion = "v2"

	// Docs: https://www.revenuecat.com/docs/api/v2#tag/Pagination
	defaultPageSize = "100"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	objectPath, err := metadata.Schemas.FindURLPath(common.ModuleRoot, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(
		c.ProviderInfo().BaseURL,
		apiVersion,
		"projects",
		c.ProjectID,
		objectPath,
	)
	if err != nil {
		return nil, err
	}

	// List endpoints support forward pagination via `limit`.
	url.WithQueryParam("limit", readhelper.PageSizeWithDefaultStr(params, defaultPageSize))

	return url, nil
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	recordsKey := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)

	return common.ParseResult(resp,
		common.ExtractOptionalRecordsFromPath(recordsKey),
		nextPageFromListObject(request.URL),
		common.GetMarshaledData,
		params.Fields,
	)
}

func nextPageFromListObject(previousRequestURL *url.URL) common.NextPageFunc {
	return func(root *ajson.Node) (string, error) {
		nextPage, err := jsonquery.New(root).StrWithDefault("next_page", "")
		if err != nil || nextPage == "" {
			return "", err
		}

		parsed, err := url.Parse(nextPage)
		if err != nil {
			return "", err
		}

		// RevenueCat usually returns a relative path (e.g. "/v2/projects/...").
		// Resolve against the *previous request URL* so we don't rely on ProviderInfo().BaseURL.
		if parsed.Scheme == "" {
			return previousRequestURL.ResolveReference(parsed).String(), nil
		}

		return nextPage, nil
	}
}
