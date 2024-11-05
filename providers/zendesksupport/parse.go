package zendesksupport

import (
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

func getNextRecordsURL(node *ajson.Node) (string, error) {
	nextPage, err := jsonquery.New(node, "links").StrWithDefault("next", "")
	if err != nil {
		return "", err
	}

	if len(nextPage) != 0 {
		return nextPage, nil
	}

	// Next page can be found under different location.
	// This format was noticed via Zendesk HelpCenter module.
	return jsonquery.New(node).StrWithDefault("next_page", "")
}
