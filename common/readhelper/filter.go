package readhelper

import (
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// FilterSortedRecords filters and returns only the records that have changed since the last sync,
// based on a provided timestamp key and reference value.
//
// Records has to be sorted, recently updated first.
//
// It compares each record's timestamp (identified by the `timestampKey`) against the provided `since` time.
// Only records with timestamps greater than the given `since` time are considered new or updated
// and included in the result.
//
// Parameters:
//   - data: A JSON node containing an array of records to filter
//   - recordsKey: The JSON path to the array of records within the data node
//   - since: The reference timestamp; only records newer than this will be returned
//   - timestampKey: The field name within each record that contains the timestamp to compare
//   - timestampFormat: The time format string used to parse timestamps from the provider
//   - nextPageFunc: Function to generate next page token if more records are available
//
// Returns:
//   - A slice of maps containing only the records updated after the specified timestamp
//   - Next page token string if more records are available, empty string if reached end
//   - An error if parsing timestamps, comparing values, or processing fails
//
// Example:
//
//	updatedRecords, nextPage, err := common.IncrementalSync(
//	    data,
//	    "records",
//	    lastSyncTime,
//	    "updated_at",
//	    time.RFC3339,
//	    nextPageFunc,
//	)
func FilterSortedRecords(data *ajson.Node, recordsKey string, since time.Time, //nolint:cyclop
	timestampKey string, timestampFormat string, nextPageFunc common.NextPageFunc,
	zoom ...string,
) ([]map[string]any, string, error) {
	var (
		updatedNodeRecords []*ajson.Node
		hasMore            bool
		next               string
	)

	nodeQuery := jsonquery.New(data)

	nodeRecords, err := nodeQuery.ArrayRequired(recordsKey)
	if err != nil {
		return nil, "", fmt.Errorf("error: bad records key: %w", err)
	}

	lastRecord := len(nodeRecords) - 1

	if len(nodeRecords) == 0 {
		return nil, next, nil
	}

	for idx, nodeRecord := range nodeRecords {
		recordTimestamp, err := extractTimestamp(nodeRecord, timestampKey, timestampFormat, zoom...)
		if err != nil {
			return nil, "", err
		}

		// Check if this record is newer than our reference time
		if since.Before(*recordTimestamp) {
			updatedNodeRecords = append(updatedNodeRecords, nodeRecord)

			// If this is the last record and it's new, we might have more pages
			if idx == lastRecord {
				hasMore = true
			}
		} else {
			// Records are assumed to be in chronological order, the function wont work otherwise.
			break
		}
	}

	if hasMore {
		next, err = nextPageFunc(data)
		if err != nil {
			return nil, next, fmt.Errorf("error: constructing next page value: %w", err)
		}
	}

	updatedRecords, err := jsonquery.Convertor.ArrayToMap(updatedNodeRecords)
	if err != nil {
		return nil, next, fmt.Errorf("error: conversion of node records to map: %w", err)
	}

	return updatedRecords, next, nil
}

// MakeIdentityFilterFunc returns a RecordsFilterFunc that allows all records to pass through unchanged.
// It delegates pagination control entirely to the provided nextPageFunc.
//
// Example usage:
//
//	return common.ParseResultFiltered(params, resp,
//		common.MakeRecordsFunc(responseFieldName),
//		readhelper.MakeIdentityFilterFunc(makeNextRecordsURL(request.URL)),
//		common.MakeMarshaledDataFunc(nil),
//		params.Fields,
//	)
func MakeIdentityFilterFunc(nextPageFunc common.NextPageFunc) common.RecordsFilterFunc {
	return func(params common.ReadParams, body *ajson.Node, records []*ajson.Node) ([]*ajson.Node, string, error) {
		next, err := nextPageFunc(body)

		return records, next, err
	}
}

// MakeTimeFilterFunc returns a RecordsFilterFunc that filters records based on timestamp boundaries.
// It uses the provided TimeOrder to determine whether pagination should continue, and TimeBoundary
// to decide whether the timestamp inclusivity applies to the Since/Until values.
//
// If records are ordered (Chronological or ReverseOrder), pagination is stopped early when all
// remaining records would fall outside the requested time range.
//
// Arguments:
//   - order: defines the chronological order of the input records.
//   - boundary: defines whether Since/Until are inclusive or exclusive.
//   - timestampKey: JSON key used to extract the timestamp value from each record.
//   - timestampFormat: time format layout for parsing timestamps.
//   - nextPageFunc: function used to determine the next page token from the response body.
func MakeTimeFilterFunc(
	order TimeOrder, boundary *TimeBoundary,
	timestampKey string, timestampFormat string,
	nextPageFunc common.NextPageFunc,
) common.RecordsFilterFunc {
	return MakeTimeFilterFuncWithZoom(order, boundary, nil, timestampKey, timestampFormat, nextPageFunc)
}

// MakeTimeFilterFuncWithZoom is similar to MakeTimeFilterFunc, but allows
// the timestamp field to be nested within the JSON record.
//
// Use `zoom` to specify the path to the nested object that contains the
// timestamp. Each element in `zoom` represents a key to traverse in order.
// Once the zoom path is resolved, `timestampKey` is read from that object
// and parsed using `timestampFormat`.
//
// Example:
//
//	If the timestamp is located at:
//	    {"meta": {"info": {"created_at": "..."} }}
//	Then zoom should be: []string{"meta", "info"}
//	And timestampKey: "created_at"
func MakeTimeFilterFuncWithZoom( // nolint:cyclop
	order TimeOrder, boundary *TimeBoundary,
	zoom []string, timestampKey string, timestampFormat string,
	nextPageFunc common.NextPageFunc,
) common.RecordsFilterFunc {
	return func(params common.ReadParams, body *ajson.Node, records []*ajson.Node) ([]*ajson.Node, string, error) {
		if len(records) == 0 {
			// Nothing to process on this page.
			return nil, "", nil
		}

		var (
			filtered   []*ajson.Node
			stopPaging bool
		)

		for _, nodeRecord := range records {
			recordTimestamp, err := extractTimestamp(nodeRecord, timestampKey, timestampFormat, zoom...)
			if err != nil {
				return nil, "", err
			}

			if boundary.Contains(params, *recordTimestamp) {
				filtered = append(filtered, nodeRecord)

				continue
			}

			switch order {
			case ChronologicalOrder:
				// [Time] oldest --> newest
				// -------------[.......]---------- * --------------->
				//           (since   UNTIL)     (record)
				// If record is after the until, then anything further is too new.
				if boundary.After(params, *recordTimestamp) {
					stopPaging = true
				}
			case ReverseOrder:
				// [Time] newest <-- oldest
				// <-------------[.......]---------- * ---------------
				//           (until   SINCE)     (record)
				// If record is before the since, then anything further is even older.
				if boundary.Before(params, *recordTimestamp) {
					stopPaging = true
				}
			case Unordered:
				// cannot infer anything
			}

			if stopPaging {
				break
			}
		}

		// Proven exhaustion.
		if stopPaging {
			return filtered, "", nil
		}

		// Continue normally.
		next, err := nextPageFunc(body)

		return filtered, next, err
	}
}

func extractTimestamp(nodeRecord *ajson.Node, timestampKey string, timestampFormat string, zoom ...string,
) (*time.Time, error) {
	// Extract the timestamp value from the record
	timestamp, err := jsonquery.New(nodeRecord, zoom...).StringRequired(timestampKey)
	if err != nil {
		return nil, fmt.Errorf("error: bad since timestamp key: %w", err)
	}

	// Parse the timestamp using the provider's specific format
	recordTimestamp, err := time.Parse(timestampFormat, timestamp)
	if err != nil {
		return nil, fmt.Errorf("error: cannot parse timestamp for key %q: %w", timestampKey, err)
	}

	return &recordTimestamp, nil
}
