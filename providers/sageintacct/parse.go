package sageintacct

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func makeNextRecordsURL(baseURL *urlbuilder.URL, params common.ReadParams) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// Look for pagination information in the response
		// Sage Intacct uses "ia::meta" for pagination metadata
		meta, err := jsonquery.New(node).ObjectOptional("ia::meta")
		if err != nil || meta == nil {
			return "", nil // No pagination info found
		}

		// Check if there is a next page
		next, err := jsonquery.New(meta).StringOptional("next")
		if err != nil || next == nil {
			return "", nil // No next page
		}

		if *next == "" {
			return "", nil // Empty next page URL means no more pages
		}

		return *next, nil
	}
}
