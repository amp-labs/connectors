package readhelper

import (
	"encoding/json"
	"log/slog"

	"github.com/amp-labs/connectors/common"
)

// AggregateNextPage holds a collection of next page tokens, each associated with a context.
// It enables state management for fetching multiple pages of the same object across different contexts.
type AggregateNextPage[C any] []NextPageToken[C]

// NextPageToken holds a next page token along with associated context.
type NextPageToken[C any] struct {
	// Context is additional state needed to resume a read operation (e.g., an account ID).
	Context C `json:"context"`
	// Value is the actual next page token for the provider.
	Value common.NextPageToken `json:"value"`
}

type (
	// PageContextCreator creates a context of type C from a row ID.
	// Used to reconstruct context when aggregating results from multiple sources.
	PageContextCreator[C any] func(rowID string) C

	// ReadResultRowModifier modifies a ReadResultRow based on its source row ID.
	// Used to inject data into result rows.
	ReadResultRowModifier func(rowID string, row *common.ReadResultRow)
)

// AggregateReadResults merges multiple ReadResults into a single ReadResult.
//
// It performs three operations:
//  1. Aggregates all rows from the registry into a single Data slice
//  2. Collects next page tokens from results that have more data, creating an AggregateNextPage
//  3. Applies the modifier function to each row for metadata injection
//
// The returned ReadResult.NextPage contains an encoded AggregateNextPage that must be
// unwrapped with GetAggregateToken when the connector receives it back for pagination.
func AggregateReadResults[C any](
	registry map[string]common.ReadResult,
	createPageContext PageContextCreator[C],
	modifyReadResultRow ReadResultRowModifier,
) *common.ReadResult {
	result := &common.ReadResult{
		Rows:     0,
		Data:     make([]common.ReadResultRow, 0),
		NextPage: "",
		Done:     false,
	}

	aggregate := make(AggregateNextPage[C], 0)
	hasMore := false

	for key, value := range registry {
		result.Rows += value.Rows
		// hasMore is true if at least one result has more data to read
		hasMore = hasMore || !(value.Done)

		if value.NextPage != "" {
			aggregate = append(aggregate, NextPageToken[C]{
				Context: createPageContext(key),
				Value:   value.NextPage,
			})
		}

		for _, row := range value.Data {
			modifyReadResultRow(key, &row) // side effect: modifies the row
			result.Data = append(result.Data, row)
		}
	}

	result.NextPage = aggregate.token()
	result.Done = !hasMore

	return result
}

// token marshals AggregateNextPage into a single common.NextPageToken.
// Returns an empty string if the aggregate is empty or if marshaling fails.
func (a AggregateNextPage[C]) token() common.NextPageToken {
	if len(a) == 0 {
		return ""
	}

	data, err := json.Marshal(a)
	if err != nil {
		slog.Warn("AggregateNextPage cannot be marshalled to create common.NextPageToken", "error", err)

		return ""
	}

	return common.NextPageToken(data)
}

// GetAggregateToken unmarshals a common.NextPageToken into an AggregateNextPage.
// Returns the aggregate and true on success, or nil and false if unmarshaling fails
// indicating that it is not an aggregate token, but a regular.
func GetAggregateToken[C any](token common.NextPageToken) (AggregateNextPage[C], bool) {
	aggregate := make(AggregateNextPage[C], 0)

	if err := json.Unmarshal([]byte(token), &aggregate); err != nil {
		// Token is not of AggregateNextPage type.
		return nil, false
	}

	return aggregate, true
}
