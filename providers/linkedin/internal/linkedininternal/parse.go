package linkedininternal

import (
	"strconv"

	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	PageSize        = 100
	CountSize       = 100
	LinkedInVersion = "202504"
	ProtocolVersion = "2.0.0"
)

func HandleCursorPagination(node *ajson.Node) (string, error) {
	pagination, err := jsonquery.New(node).ObjectOptional("metadata")
	if err != nil {
		return "", err
	}

	if pagination != nil {
		nextPage, err := jsonquery.New(pagination).StrWithDefault("nextPageToken", "")
		if err != nil {
			return "", err
		}

		if nextPage != "" {
			return nextPage, nil
		}
	}

	return "", nil
}

func HandleOffsetPagination(node *ajson.Node) (string, error) {
	paging, err := jsonquery.New(node).ObjectOptional("paging")
	if err != nil {
		return "", err
	}

	if paging != nil {
		nextPage, err := jsonquery.New(paging).IntegerWithDefault("count", 0)
		if err != nil {
			return "", err
		}

		if nextPage != 0 {
			start, err := jsonquery.New(paging).IntegerWithDefault("start", 0)
			if err != nil {
				return "", err
			}

			return strconv.Itoa(int(start) + int(nextPage)), nil
		}
	}

	return "", nil
}
