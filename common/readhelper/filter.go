package readhelper

import (
	"errors"
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

var ErrKeyNotFound = errors.New("incsync: key not found in one or more records; please verify the key")

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
		// Extract the timestamp value from the record
		timestamp, err := jsonquery.New(nodeRecord).StringRequired(timestampKey)
		if err != nil {
			return nil, next, fmt.Errorf("error: bad since timestamp key: %w", err)
		}

		// Parse the timestamp using the provider's specific format
		recordTimestamp, err := time.Parse(timestampFormat, timestamp)
		if err != nil {
			return nil, next, fmt.Errorf("error: cannot parse timestamp for key %q: %w", timestampKey, err)
		}

		// Check if this record is newer than our reference time
		if since.Before(recordTimestamp) {
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
