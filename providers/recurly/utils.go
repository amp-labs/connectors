package recurly

import (
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func buildNextPageURL(baseURL, nextPage string) (*urlbuilder.URL, error) {
	// NextPage contains the full path with cursor,
	// e.g., "/accounts?cursor=xy0togeu9vun%3A1763384298.485542&limit=2&sort=created_at"
	return urlbuilder.New(baseURL + nextPage)
}

func buildFirstPageURL(baseURL string, params common.ReadParams) (*urlbuilder.URL, error) {
	url, err := urlbuilder.New(baseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", limit)
	addTimeFilters(url, params)

	return url, nil
}

func addTimeFilters(url *urlbuilder.URL, params common.ReadParams) {
	if !supportIncrementalRead.Has(params.ObjectName) {
		return
	}

	if !params.Since.IsZero() {
		url.WithQueryParam("begin_time", params.Since.Format(time.RFC3339))
	}

	if !params.Until.IsZero() {
		url.WithQueryParam("end_time", params.Until.Format(time.RFC3339))
	}
}
