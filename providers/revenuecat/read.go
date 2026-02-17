package revenuecat

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/amp-labs/connectors/common"
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
		next := params.NextPage.String()

		// `next_page` may be relative; anchor it on BaseURL.
		if parsed, err := url.Parse(next); err == nil && parsed.Scheme == "" {
			base, err2 := url.Parse(c.ProviderInfo().BaseURL)
			if err2 != nil {
				return nil, err2
			}
			return urlbuilder.New(base.ResolveReference(parsed).String())
		}

		return urlbuilder.New(next)
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
	limit := defaultPageSize
	if params.PageSize > 0 {
		limit = strconv.Itoa(params.PageSize)
	}

	url.WithQueryParam("limit", limit)

	return url, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	_ = ctx

	recordsKey := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)

	return common.ParseResult(resp,
		common.ExtractOptionalRecordsFromPath(recordsKey),
		nextPageFromListObject(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func nextPageFromListObject() common.NextPageFunc {
	return func(root *ajson.Node) (string, error) {
		// `next_page` is present only when more pages exist.
		next, err := jsonquery.New(root).StrWithDefault("next_page", "")
		if err != nil {
			return "", err
		}

		return next, nil
	}
}
