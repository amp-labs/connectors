package intercom

import (
	"errors"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/gertd/go-pluralize"
	"github.com/spyzhov/ajson"
)

/*
Response example:

	{
	  "type": "list",
	  "data": [{...}],
	  "total_count": 1,
	  "pages": {
		"type": "pages",
		"page": 1,
		"next": "https://api.intercom.io/contacts/6643703ffae7834d1792fd30/notes?per_page=1&page=2",
		"per_page": 100,
		"total_pages": 1
	  }
	}

Note:

	=> `pages.next` can be null.
	=> Sometimes array of objects is not stored at `data` but named after `type`.
*/
func getRecords(node *ajson.Node) ([]map[string]any, error) {
	arrKey, err := extractListFieldName(node)
	if err != nil {
		return nil, err
	}

	arr, err := jsonquery.New(node).ArrayRequired(arrKey)
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(arr)
}

func makeNextRecordsURL(reqLink *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		next, err := getNextPageStringURL(node)
		if err == nil {
			return next, nil
		}

		if !errors.Is(err, jsonquery.ErrNotString) {
			// response from server doesn't meet any format that we expect
			return "", err
		}

		// Probably, we are dealing with an object under `pages.next`
		startingAfter, err := jsonquery.New(node, "pages", "next").StringOptional("starting_after")
		if err != nil {
			return "", err
		}

		if startingAfter == nil {
			// next page doesn't exist
			return "", nil
		}

		reqLink.WithQueryParam("starting_after", *startingAfter)

		return reqLink.String(), nil
	}
}

// Some responses have full URL stored at `pages.next`.
func getNextPageStringURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node, "pages").StrWithDefault("next", "")
}

// Intercom returns a type field which tells us where the response array is located,
// which is why we do not need to hardcode the mapping. If we need to override this at any point,
// we can add the mapping here.
//
// Other connectors don't have the ability to infer field names programmatically, so they rely on hardcoded mappings.
// In this case, the response field name will be dynamically determined using the value of the "type" field.
// Ex:
//
//	{"type":"data", "data": []}
//	{"type":"teams", "teams":[]}
//	{"type":"segments", "segments":[]}
func extractListFieldName(node *ajson.Node) (string, error) {
	// default field at which list is stored
	defaultFieldName := "data"

	fieldName, err := jsonquery.New(node).StringOptional("type")
	if err != nil {
		return "", err
	}

	if fieldName == nil {
		// this object has no `type` field to infer where the array is situated
		// it is unexpected to encounter it
		return defaultFieldName, nil
	}

	name := *fieldName
	// by applying plural form to the object name we will the name of field containing array
	// Ex with `list` suffix:
	// 		activity_log.list => activity_logs
	// 		admin.list => admins
	// 		conversation.list => conversations
	// 		segment.list => segments
	// 		team.list => teams
	// Exceptions:
	//		event.summary => events

	parts := strings.Split(name, ".")
	if len(parts) == 2 { // nolint:mnd
		// custom name is used when it has 2 parts
		return applyPluralForm(parts[0]), nil
	}

	// usually when we have a pure `list` type it means array is stored at `data` field
	return defaultFieldName, nil
}

func applyPluralForm(word string) string {
	return pluralize.NewClient().Plural(word)
}
