package atlassian

import (
	"errors"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

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
func flattenRecord(item *ajson.Node) (map[string]any, error) {
	fieldsObject, err := jsonquery.New(item).ObjectRequired("fields")
	if err != nil {
		return nil, errors.Join(common.ErrParseError, err)
	}

	fields, err := jsonquery.Convertor.ObjectToMap(fieldsObject)
	if err != nil {
		return nil, errors.Join(common.ErrParseError, err)
	}

	id, err := jsonquery.New(item).StringRequired("id")
	if err != nil {
		return nil, errors.Join(common.ErrParseError, err)
	}

	// Enhance response with id property.
	fields["id"] = id

	return fields, nil
}

func getRecords(node *ajson.Node) ([]*ajson.Node, error) {
	return jsonquery.New(node).ArrayRequired("issues")
}

// Next starting page index is calculated base on current index and array size.
func getNextRecords(node *ajson.Node) (string, error) {
	records, err := getRecords(node)
	if err != nil {
		return "", err
	}

	size := int64(len(records))

	if size == 0 {
		// No elements returned for the current page.
		// There is no need to go further, definitely we are at the end.
		return "", nil
	}

	startAt, err := jsonquery.New(node).IntegerOptional("startAt")
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

// Next starting page index is calculated base on current index and array size.
func getNextRecordIssues(node *ajson.Node) (string, error) {
	q := jsonquery.New(node)

	nextPageToken, err := q.StringOptional("nextPageToken")
	if err != nil {
		return "", err
	}

	isLast, err := q.BoolOptional("isLast")
	if err != nil {
		return "", err
	}

	if isLast != nil && *isLast {
		return "", nil
	}

	if nextPageToken == nil {
		return "", nil
	}

	return *nextPageToken, nil
}
