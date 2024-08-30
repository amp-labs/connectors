package gong

import (
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

// getNextRecords returns the token or empty string if there are no more records.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node, "records").StrWithDefault("cursor", "")
}
