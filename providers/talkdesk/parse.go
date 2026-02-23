package talkdesk

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func getRecords(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		keys := responseKeys[objectName]
		if len(keys) > 1 {
			rcds, err := jsonquery.New(node, keys[0]).ArrayOptional(keys[1])
			if err != nil {
				return nil, err
			}

			return jsonquery.Convertor.ArrayToMap(rcds)
		}

		rcds, err := jsonquery.New(node).ArrayOptional("_embedded")
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(rcds)
	}
}

func nextRecordsURL(root *ajson.Node) (string, error) {
	next, err := jsonquery.New(root, "_links").ObjectOptional("next")
	if err != nil {
		return "", err
	}

	if next == nil {
		return "", nil
	}

	url, err := jsonquery.New(next).StringOptional("href")
	if err != nil {
		return "", err
	}

	if url == nil {
		return "", nil
	}

	return *url, nil
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{"*"}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
