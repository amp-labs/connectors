package avoma

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion = "v1"
	pageSize   = "100"
)

// avoma pagination cursor sometimes ends with `=`.
var avomaQueryEncodingExceptions = map[string]string{ //nolint:gochecknoglobals
	"%3A": ":",
}

var EndpointsWithResultsPath = datautils.NewSet( //nolint:gochecknoglobals
	"meetings",
	"calls",
	"custom_categories",
	"notes",
	"scorecard_evaluations",
	"smart_categories",
)

func makeNextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextPage, err := jsonquery.New(node).StringRequired("next")
		if err != nil {
			return "", nil //nolint:nilerr
		}

		url, err := urlbuilder.New(nextPage)
		if err != nil {
			return "", err
		}

		url.AddEncodingExceptions(avomaQueryEncodingExceptions)

		return url.String(), nil
	}
}
