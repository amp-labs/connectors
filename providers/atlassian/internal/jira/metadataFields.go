package jira

import (
	"errors"

	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

var (
	ErrParsingMetadata = errors.New("couldn't parse metadata")
	ErrMissingMetadata = errors.New("there is no metadata for object")
)

// Converts API response into the fields' registry.
func (a *Adapter) parseFieldsJiraIssue(node *ajson.Node) (map[string]string, error) {
	arr, err := node.GetArray()
	if err != nil {
		return nil, err
	}

	if len(arr) == 0 {
		return nil, ErrMissingMetadata
	}

	fieldsMap := make(map[string]string)

	for _, item := range arr {
		name, err := jsonquery.New(item).StringRequired("id")
		if err != nil {
			return nil, err
		}

		displayName, err := jsonquery.New(item).StringRequired("name")
		if err != nil {
			return nil, err
		}

		fieldsMap[name] = displayName
	}

	return fieldsMap, nil
}
