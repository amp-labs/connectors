package ashby

import (
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

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
