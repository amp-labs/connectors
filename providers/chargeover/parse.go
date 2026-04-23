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

func nextRecordsURL(url *urlbuilder.URL, objectName string, numRecords int) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		var prevOffsetStr string

		if numRecords < pageSize {
			return "", nil
		}

		if doNotPaginate.Has(objectName) {
			return "", nil
		}

		prevOffsetStr, found := url.GetFirstQueryParam("offset")
		if !found {
			prevOffsetStr = "0"
		}

		prevOffset, err := strconv.Atoi(prevOffsetStr)
		if err != nil {
			return "", err
		}

		nextOffset := pageSize + prevOffset

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

// doNotFilter stores a list of objects that do not accept filtering when reading.
var doNotFilter = datautils.NewSet("brand", "_report", "campaign", "class", "country", "coupon", "currency", //nolint: gochecknoglobals,lll
	"_resthook", "language", "terms", "_customfield", "_log_system", "_chargeoverjs",
)

// doNotPaginate stores a list of objects whose endpoints do not take pagination parameters.
var doNotPaginate = datautils.NewSet("_report", "_chargeoverjs", "_log_system") //nolint: gochecknoglobals

var filteringFields = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint: gochecknoglobals
	"customer":     "mod_datetime",
	"invoice":      "mod_datetime",
	"transaction":  "transaction_datetime",
	"quote":        "date",
	"package":      "mod_datetime",
	"usage":        "from",
	"_customfield": "date",
}, func(key string) string {
	return ""
})
