package fireflies

import (
	"regexp"
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

// For object names like activeMeetings or userGroups, we need to display them as
// "Active Meetings" or "User Groups".
// This code inserts a space between a lowercase letter followed by an uppercase letter,
// effectively splitting a camelCase word into separate words.
func createDisplayName(objName string) string {
	re := regexp.MustCompile(`([a-z])([A-Z])`)

	return re.ReplaceAllString(objName, `${1} ${2}`)
}
