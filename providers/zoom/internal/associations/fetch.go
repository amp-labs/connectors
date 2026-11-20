package associations

import (
	"errors"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// readFunc adds the records of one response to records, and returns the URL of the next page, or
// "" when there is none.
type readFunc func(rsp common.BatchResponse, records recordItems) (string, error)

// readItem treats the entire JSON response as one record.
func readItem(identifierField string) readFunc {
	return func(rsp common.BatchResponse, records recordItems) (string, error) {
		if rsp.Err != nil {
			return "", rsp.Err
		}

		body, ok := rsp.Response.Body()
		if !ok {
			return "", nil
		}

		record, err := jsonquery.Convertor.ObjectToMap(body)
		if err != nil {
			return "", err
		}

		if len(record) == 0 {
			return "", nil
		}

		identifier, err := jsonquery.New(body).TextWithDefault(identifierField, "")
		if err != nil {
			return "", err
		}

		records[identifier] = record

		return "", nil
	}
}

// readItems reads one page of a paginated collection.
func readItems(responseFieldName, identifierField string) readFunc {
	return func(rsp common.BatchResponse, records recordItems) (string, error) {
		if rsp.Err != nil {
			return "", rsp.Err
		}

		body, ok := rsp.Response.Body()
		if !ok {
			return "", nil
		}

		array, err := jsonquery.New(body).ArrayOptional(responseFieldName)
		if err != nil {
			return "", err
		}

		if len(array) == 0 {
			return "", nil
		}

		if err := addRecords(array, identifierField, records); err != nil {
			return "", err
		}

		nextPageToken, err := jsonquery.New(body).StrWithDefault("next_page_token", "")
		if err != nil {
			return "", err
		}

		if nextPageToken == "" {
			return "", nil
		}

		url, err := urlbuilder.New(rsp.Request.URL)
		if err != nil {
			return "", err
		}

		url.WithQueryParam("next_page_token", nextPageToken)

		return url.String(), nil
	}
}

// addRecords adds every item of a page to records, keyed by its identifier.
func addRecords(items []*ajson.Node, identifierField string, records recordItems) error {
	for _, item := range items {
		record, err := jsonquery.Convertor.ObjectToMap(item)
		if err != nil {
			return err
		}

		identifier, err := jsonquery.New(item).TextWithDefault(identifierField, "")
		if err != nil {
			return err
		}

		records[identifier] = record
	}

	return nil
}

// readTranscript reads a meeting transcript. A meeting without one has no records.
func readTranscript(rsp common.BatchResponse, records recordItems) (string, error) {
	if rsp.Err != nil && errors.Is(rsp.Err, common.ErrNotFound) && strings.Contains(rsp.Err.Error(), "error code (3322)") {
		// This meeting transcript does not exist.
		// No associations to return.
		return "", nil
	}

	return readItem("meeting_id")(rsp, records)
}
