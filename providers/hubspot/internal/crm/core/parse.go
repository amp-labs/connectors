package core

import (
	"errors"
	"strconv"

	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

var (
	ErrNotArray  = errors.New("results is not an array")
	ErrNotObject = errors.New("result is not an object")
	ErrNotString = errors.New("link is not a string")
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
	var nextPage string

	if node.HasKey("paging") {
		next, err := parsePagingNext(node)
		if err != nil {
			return "", err
		}

		after, err := next.GetKey("after")
		if err != nil {
			return "", err
		}

		if !after.IsString() {
			return "", ErrNotString
		}

		nextPage = after.MustString()
	}

	return nextPage, nil
}

// GetNextRecordsURL returns the URL for the next page of results.
func GetNextRecordsURL(node *ajson.Node) (string, error) {
	var nextPage string

	if node.HasKey("paging") {
		next, err := parsePagingNext(node)
		if err != nil {
			return "", err
		}

		link, err := next.GetKey("link")
		if err != nil {
			return "", err
		}

		if !link.IsString() {
			return "", ErrNotString
		}

		nextPage = link.MustString()
	}

	return nextPage, nil
}

// parsePagingNext is a helper to return the paging.next node.
func parsePagingNext(node *ajson.Node) (*ajson.Node, error) {
	paging, err := node.GetKey("paging")
	if err != nil {
		return nil, err
	}

	if !paging.IsObject() {
		return nil, ErrNotObject
	}

	next, err := paging.GetKey("next")
	if err != nil {
		return nil, err
	}

	if !next.IsObject() {
		return nil, ErrNotObject
	}

	return next, nil
}

// GetRecords returns the records from the response.
func GetRecords(node *ajson.Node) ([]map[string]any, error) {
	records, err := node.GetKey("results")
	if err != nil {
		return nil, err
	}

	if !records.IsArray() {
		return nil, ErrNotArray
	}

	arr := records.MustArray()

	out := make([]map[string]any, 0, len(arr))

	for _, v := range arr {
		if !v.IsObject() {
			return nil, ErrNotObject
		}

		data, err := v.Unpack()
		if err != nil {
			return nil, err
		}

		m, ok := data.(map[string]any)
		if !ok {
			return nil, ErrNotObject
		}

		out = append(out, m)
	}

	return out, nil
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
