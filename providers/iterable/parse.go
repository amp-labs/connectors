package iterable

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/providers/iterable/metadata"
	"github.com/spyzhov/ajson"
)

func makeGetRecords(moduleID common.ModuleID, objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		responseFieldName := metadata.Schemas.LookupArrayFieldName(moduleID, objectName)

		var nestedPath []string
		if objectName == objectNameCatalogs {
			// For some reason this is the only object that is nested.
			// For reference, the parent object is "IterableApiResponse", which can be found under OpenAPI.
			nestedPath = []string{"params"}
		}

		return common.GetOptionalRecordsUnderJSONPath(responseFieldName, nestedPath...)(node)
	}
}

func makeNextRecordsURL(baseURL string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextPage, err := jsonquery.New(node).StrWithDefault("nextPageUrl", "")
		if err != nil {
			return "", err
		}

		if len(nextPage) == 0 {
			// Next page URL could be nested under params object.
			nextPage, err = jsonquery.New(node, "params").StrWithDefault("nextPageUrl", "")
			if err != nil {
				return "", err
			}
		}

		if len(nextPage) == 0 {
			// Next page doesn't exist
			return "", nil
		}

		fullURL := baseURL + nextPage

		return fullURL, nil
	}
}
