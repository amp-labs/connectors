package ashby

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/ashby/metadata"
	"github.com/spyzhov/ajson"
)

func getRecords(objectName string, moduleID common.ModuleID) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		responseKey := metadata.Schemas.LookupArrayFieldName(moduleID, objectName)

		rcds, err := jsonquery.New(node).ArrayOptional(responseKey)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(rcds)
	}
}

func makeNextRecordsURL(node *ajson.Node) (string, error) {
	moreDataAvailable, err := jsonquery.New(node).BoolWithDefault("moreDataAvailable", true)
	if !moreDataAvailable || err != nil {
		return "", nil //nolint:nilerr
	}

	cursor, err := jsonquery.New(node).StringOptional("nextCursor")
	if err != nil {
		return "", nil //nolint:nilerr
	}

	return *cursor, nil
}
