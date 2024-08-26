package atlassian

import (
	"errors"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

func getTotalSize(node *ajson.Node) (int64, error) {
	return jsonquery.New(node).ArraySize("issues")
}

/*
Records cannot be returned as is from the API. Extra processing is described below.
 1. First of all, main properties are located under "fields" key.
 2. Secondly, item id is not mirrored under "fields" property, which means
    unwrapping is not enough, we must attach "id" property.

Visual example of what will happen to each property:

	{
		"id": "", 					=> preserved
		"expand": "",				=> removed
		"self": "",					=> removed
		"key": "",					=> removed
		"fields": {					=> unwrapped/flattened
			"project": "",			=> moved outside
			"lastViewed": ""		=> moved outside
			...						...
			...						=> moved outside
		}
	}
*/
func getRecords(node *ajson.Node) ([]map[string]any, error) {
	arr, err := jsonquery.New(node).Array("issues", false)
	if err != nil {
		return nil, err
	}

	list := make([]map[string]any, len(arr))

	for index, item := range arr {
		fieldsObject, err := jsonquery.New(item).Object("fields", false)
		if err != nil {
			return nil, errors.Join(common.ErrParseError, err)
		}

		fields, err := jsonquery.Convertor.ObjectToMap(fieldsObject)
		if err != nil {
			return nil, errors.Join(common.ErrParseError, err)
		}

		id, err := jsonquery.New(item).Str("id", false)
		if err != nil {
			return nil, errors.Join(common.ErrParseError, err)
		}

		// Enhance response with id property.
		fields["id"] = *id
		list[index] = fields
	}

	return list, nil
}

// Next starting page index is calculated base on current index and array size.
func getNextRecords(node *ajson.Node) (string, error) {
	size, err := getTotalSize(node)
	if err != nil {
		return "", err
	}

	if size == 0 {
		// No elements returned for the current page.
		// There is no need to go further, definitely we are at the end.
		return "", nil
	}

	startAt, err := jsonquery.New(node).Integer("startAt", true)
	if err != nil {
		return "", err
	}

	if startAt == nil {
		// we cannot determine the next page
		return "", nil
	}

	// StartAt starts from zero
	nextStartIndex := *startAt + size

	return strconv.FormatInt(nextStartIndex, 10), nil
}
