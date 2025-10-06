package gong

import (
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// getNextRecords returns the token or empty string if there are no more records.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node, "records").StrWithDefault("cursor", "")
}

func getRecords(responseKey string) func(node *ajson.Node) ([]*ajson.Node, error) {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayRequired(responseKey)
	}
}
