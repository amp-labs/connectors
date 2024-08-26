package atlassian

import (
	"errors"

	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

var (
	ErrParsingMetadata = errors.New("couldn't parse metadata")
	ErrMissingMetadata = errors.New("there is no metadata for object")
)

// Converts API response into the fields' registry.
func (c *Connector) parseFieldsJiraIssue(node *ajson.Node) (map[string]string, error) {
	arr, err := node.GetArray()
	if err != nil {
		return nil, err
	}

	if len(arr) == 0 {
		return nil, ErrMissingMetadata
	}

	fieldsMap := make(map[string]string)

	for _, item := range arr {
		name, err := jsonquery.New(item).Str("id", false)
		if err != nil {
			return nil, err
		}

		displayName, err := jsonquery.New(item).Str("name", false)
		if err != nil {
			return nil, err
		}

		fieldsMap[*name] = *displayName
	}

	return fieldsMap, nil
}
