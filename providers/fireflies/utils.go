package fireflies

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	defaultPageSize = 50
)

var supportLimitAndSkip = datautils.NewSet( //nolint:gochecknoglobals
	"transcripts",
	"bites",
)

var objectNameMapping = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"userGroups":     "user_groups",
	"activeMeetings": "active_meetings",
}, func(objectName string) string {
	return objectName
})

func makeNextRecordsURL(params common.ReadParams) func(*ajson.Node) (string, error) {
	return func(node *ajson.Node) (string, error) {
		if !supportLimitAndSkip.Has(params.ObjectName) {
			return "", nil
		}

		records, err := jsonquery.New(node, "data").ArrayRequired(objectNameMapping.Get(params.ObjectName))
		if err != nil {
			return "", err
		}

		if len(records) < defaultPageSize {
			return "", nil
		}

		var currentPage int

		if params.NextPage != "" {
			currentPage, err = strconv.Atoi(params.NextPage.String())
			if err != nil {
				return "", err
			}
		}

		nextPage := currentPage + len(records)

		return strconv.Itoa(nextPage), nil
	}
}
