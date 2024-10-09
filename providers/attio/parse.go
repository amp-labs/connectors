package attio

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

func makeNextRecordsURL(reqLink *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		value := node.MustObject()["data"].MustArray()
		previousStart := 0
		url, err := reqLink.ToURL()
		if err != nil {
			return "", err
		}

		if (url.Query().Has("limit") || url.Query().Has("offset")) && len(value) != 0 {

			offsetQP, ok := reqLink.GetFirstQueryParam("offset")
			if ok {
				// Try to use previous "offset" parameter to determine the next offset.
				offsetNum, err := strconv.Atoi(offsetQP)
				if err == nil {
					previousStart = offsetNum
				}
			}

			nextStart := previousStart + DefaultPageSize

			reqLink.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))
			reqLink.WithQueryParam("offset", strconv.Itoa(nextStart))

			return reqLink.String(), nil

		} else {
			return "", nil
		}
	}
}
