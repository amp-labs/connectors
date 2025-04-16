package zendesksupport

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
	"github.com/spyzhov/ajson"
)

// https://developer.zendesk.com/api-reference/ticketing/ticket-management/incremental_exports/#json-format
func getNextRecordsURL(node *ajson.Node) (string, error) {
	isStreamEnd, err := jsonquery.New(node).BoolWithDefault("end_of_stream", false)
	if err != nil {
		return "", err
	}

	if isStreamEnd {
		// Time-based pagination would still return next page even thought he next page doesn't exist.
		// We must stop paginating if the end of stream is reached.
		return "", nil
	}

	hasMore, err := jsonquery.New(node, "meta").BoolWithDefault("has_more", true)
	if err != nil {
		return "", err
	}

	if !hasMore {
		return "", nil
	}

	nextPage, err := jsonquery.New(node, "links").StrWithDefault("next", "")
	if err != nil {
		return "", err
	}

	if len(nextPage) != 0 {
		return nextPage, nil
	}

	// Next page can be found under different location.
	// This format was noticed via Zendesk HelpCenter module.
	nextPage, err = jsonquery.New(node).StrWithDefault("next_page", "")
	if err != nil {
		return "", err
	}

	if len(nextPage) != 0 {
		return nextPage, nil
	}

	return jsonquery.New(node).StrWithDefault("after_url", "")
}

func getRecords(moduleID common.ModuleID, objectName string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		responseFieldName := metadata.Schemas.LookupArrayFieldName(moduleID, objectName)

		return jsonquery.New(node).ArrayRequired(responseFieldName)
	}
}
