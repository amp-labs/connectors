package sybill

import (
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// nextPage returns pagination.nextCursor while pagination.hasMore is true.
// The docs say to check hasMore, not nextCursor, to tell whether more pages exist.
// https://api.sybill.ai/docs/pagination.html#best-practices
func nextPage(node *ajson.Node) (string, error) {
	query := jsonquery.New(node, "pagination")

	hasMore, err := query.BoolWithDefault("hasMore", false)
	if err != nil || !hasMore {
		return "", err
	}

	cursor, err := query.StringOptional("nextCursor")
	if err != nil || cursor == nil {
		return "", err
	}

	return *cursor, nil
}
