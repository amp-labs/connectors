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

var endpointsWithResultsPath = datautils.NewSet( //nolint:gochecknoglobals
	"meetings",
	"calls",
	"notes",
	"scorecard_evaluations",
)

func makeNextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextPage, err := jsonquery.New(node).StringOptional("next")
		if err != nil {
			return "", err
		}

		if nextPage == nil {
			return "", nil
		}

		url, err := urlbuilder.New(*nextPage)
		if err != nil {
			return "", err
		}

		url.AddEncodingExceptions(avomaQueryEncodingExceptions)

		return url.String(), nil
	}
}
