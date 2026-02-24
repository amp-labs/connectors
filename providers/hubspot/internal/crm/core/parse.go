package core

import (
	"errors"
	"fmt"
	"maps"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

var (
	ErrNotObject = errors.New("result is not an object")
	ErrMissingId = errors.New("missing id field in raw record")
)

/*
Pagination format:

{
  "results": [...],
  "paging": {
    "next": {
      "after": "394",
      "link": "https://api.hubapi.com/crm/v3/objects/contacts?limit=100&properties=listId%2Cname&after=394"
    }
  }
}
*/

// GetNextRecordsAfter returns the "after" value for the next page of results.
func GetNextRecordsAfter(node *ajson.Node) (string, error) {
	return jsonquery.New(node, "paging", "next").StrWithDefault("after", "")
}

// GetNextRecordsURL returns the URL for the next page of results.
func GetNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node, "paging", "next").StrWithDefault("link", "")
}

// GetRecords returns the records from the response.
func GetRecords(node *ajson.Node) ([]map[string]any, error) {
	extractor := common.ExtractRecordsFromPath("results")

	return extractor(node)
}

func GetNextRecordsURLCRM(node *ajson.Node) (string, error) {
	hasMore, err := jsonquery.New(node).BoolWithDefault("hasMore", false)
	if err != nil {
		return "", err
	}

	if !hasMore {
		// Next page doesn't exist
		return "", nil
	}

	offset, err := jsonquery.New(node).IntegerWithDefault("offset", 0)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(offset, 10), nil
}

// GetDataMarshaller returns a function that accepts a list of records and fields
// and returns a list of structured data ([]ReadResultRow).
//
//nolint:gocognit
func GetDataMarshaller() common.MarshalFromNodeFunc {
	return readhelper.MakeMarshaledSelectedDataFunc(
		fieldsFromProperties(),
		func(node *ajson.Node) (map[string]any, error) {
			return jsonquery.Convertor.ObjectToMap(node)
		},
	)
}

func fieldsFromProperties() readhelper.SelectedFieldsFunc {
	return func(node *ajson.Node, fields []string) (map[string]any, string, error) {
		root, err := jsonquery.Convertor.ObjectToMap(node)
		if err != nil {
			return nil, "", err
		}

		identifier, err := jsonquery.New(node).StringRequired("id")
		if err != nil {
			return nil, "", fmt.Errorf("missing id field in raw record: %w", err)
		}

		properties, err := jsonquery.New(node).ObjectRequired("properties")
		if err != nil {
			return nil, "", err
		}

		propertiesMap, err := jsonquery.Convertor.ObjectToMap(properties)
		if err != nil {
			return nil, identifier, err
		}

		filteredRoot := readhelper.SelectFields(root, datautils.NewSetFromList(fields))
		selected := readhelper.SelectFields(propertiesMap, datautils.NewSetFromList(fields))
		maps.Copy(filteredRoot, selected)

		return filteredRoot, identifier, nil
	}
}
