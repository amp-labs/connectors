package aweber

import (
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

func makeNextRecordsURL(url *urlbuilder.URL, header http.Header) common.NextPageFunc {
	return func(_ *ajson.Node) (string, error) {
		currentPage := header.Get("CurrentPage")

		url.WithQueryParam("page", currentPage)

		return url.String(), nil
	}
}
