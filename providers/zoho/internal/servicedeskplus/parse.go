package servicedeskplus

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

var hasCreationTime = datautils.NewStringSet( //nolint: gochecknoglobals
	"requests", "problems", "changes", "projects",
	"releases", "assets", "solutions", "space_campuses",
	"space_buildings", "space_nonbuildings", "space_floors",
	"space_rooms", "space_roompartitions",
)

var hasCreationDate = datautils.NewStringSet( //nolint: gochecknoglobals
	"tasks", "purchase_orders", "announcements",
)

func extractRecordsFromPath(objectName string) common.RecordsFunc {
	return common.ExtractRecordsFromPath(objectName)
}

func getNextRecordsURL(node *ajson.Node) (string, error) {
	hasMoreRecords, err := jsonquery.New(node, "list_info").BoolWithDefault("has_more_rows", false)
	if err != nil {
		return "", err
	}

	if hasMoreRecords {
		currentPage, err := jsonquery.New(node, "list_info").IntegerOptional("page")
		if err != nil {
			return "", err
		}

		if currentPage != nil {
			return strconv.Itoa(int(*currentPage + 1)), nil
		}
	}

	return "", nil
}
