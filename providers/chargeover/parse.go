package chargeover

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL(url *urlbuilder.URL, numRecords, offset int) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		if numRecords < pageSize {
			return "", nil
		}

		nextOffset := pageSize + offset + 1

		url.WithQueryParam(offsetQuery, strconv.Itoa(nextOffset))

		return url.String(), nil
	}
}

func supportedOperations() components.EndpointRegistryInput { //nolint:funlen
	readSupport := []string{"*"}

	writeSupport := []string{"*"}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}

var doNotFilter = datautils.NewSet("brand", "_report", "campaign", "class", "country", "coupon", "currency", //nolint: gochecknoglobals,lll
	"_resthook", "language", "terms", "_customfield", "_log_system", "_chargeoverjs",
)
