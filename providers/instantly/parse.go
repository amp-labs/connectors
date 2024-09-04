package instantly

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

func makeNextRecordsURL(reqLink *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		previousStart := 0

		skipQP, ok := reqLink.GetFirstQueryParam("skip")
		if ok {
			// Try to use previous "skip" parameter to determine the next skip.
			skipNum, err := strconv.Atoi(skipQP)
			if err == nil {
				previousStart = skipNum
			}
		}

		nextStart := previousStart + DefaultPageSize

		reqLink.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))
		reqLink.WithQueryParam("skip", strconv.Itoa(nextStart))

		return reqLink.String(), nil
	}
}
