// nolint
package readhelper

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/spyzhov/ajson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_FilterSortedRecords(t *testing.T) {
	// Create test data helper
	createTestData := func(records []map[string]any) *ajson.Node {
		jsonBytes, err := json.Marshal(map[string]any{"records": records})
		require.NoError(t, err)

		node, err := ajson.Unmarshal(jsonBytes)
		require.NoError(t, err)

		return node
	}

	createRecord := func(id string, updatedAt string) map[string]any {
		return map[string]any{
			"id":         id,
			"updated_at": updatedAt,
			"name":       "record-" + id,
		}
	}

	mockNextPageFunc := func(*ajson.Node) (string, error) {
		return "next-page", nil
	}

	errorNextPageFunc := func(*ajson.Node) (string, error) {
		return "", errors.New("next page error")
	}

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{"empty records array", testEmptyRecords(createTestData, mockNextPageFunc)},
		{"all records are newer than since time", testAllRecordsNewer(createTestData, createRecord, mockNextPageFunc)},
		{"some records are newer, some older", testMixedRecords(createTestData, createRecord, mockNextPageFunc)},
		{"all records are older than since time", testAllRecordsOlder(createTestData, createRecord, mockNextPageFunc)},
		{"last record is newer - should indicate more pages", testLastRecordNewer(createTestData, createRecord, mockNextPageFunc)},
		{"invalid records key", testInvalidRecordsKey(createTestData, createRecord, mockNextPageFunc)},
		{"invalid timestamp key", testInvalidTimestampKey(createTestData, createRecord, mockNextPageFunc)},
		{"invalid timestamp format", testInvalidTimestampFormat(createTestData, mockNextPageFunc)},
		{"next page function returns error", testNextPageError(createTestData, createRecord, errorNextPageFunc)},
		{"different timestamp format", testDifferentTimestampFormat(createTestData, mockNextPageFunc)},
		{"records with exact same timestamp as since", testExactSameTimestamp(createTestData, createRecord, mockNextPageFunc)},
		{"complex nested JSON structure", testComplexNestedJSON(mockNextPageFunc)},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

func testEmptyRecords(createTestData func([]map[string]any) *ajson.Node, mockNextPageFunc func(*ajson.Node) (string, error),
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		data := createTestData([]map[string]any{})
		since := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		records, nextPage, err := FilterSortedRecords(
			data, "records", since, "updated_at", time.RFC3339, mockNextPageFunc,
		)

		assert.NoError(t, err)
		assert.Empty(t, records)
		assert.Empty(t, nextPage)
	}
}

func testAllRecordsNewer(createTestData func([]map[string]any) *ajson.Node,
	createRecord func(string, string) map[string]any, mockNextPageFunc func(*ajson.Node) (string, error),
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		records := []map[string]any{
			createRecord("1", "2023-01-03T10:00:00Z"),
			createRecord("2", "2023-01-02T10:00:00Z"),
			createRecord("3", "2023-01-01T10:00:00Z"),
		}
		data := createTestData(records)
		since := time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC)

		result, nextPage, err := FilterSortedRecords(
			data, "records", since, "updated_at", time.RFC3339, mockNextPageFunc,
		)

		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, "next-page", nextPage)
		assert.Equal(t, "1", result[0]["id"])
		assert.Equal(t, "2", result[1]["id"])
		assert.Equal(t, "3", result[2]["id"])
	}
}

func testMixedRecords(createTestData func([]map[string]any) *ajson.Node,
	createRecord func(string, string) map[string]any, mockNextPageFunc func(*ajson.Node) (string, error),
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		records := []map[string]any{
			createRecord("1", "2023-01-03T10:00:00Z"), // newer
			createRecord("2", "2023-01-02T10:00:00Z"), // newer
			createRecord("3", "2022-12-31T10:00:00Z"), // older - should stop here
			createRecord("4", "2023-01-01T10:00:00Z"), // should not be processed due to break
		}
		data := createTestData(records)
		since := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		result, nextPage, err := FilterSortedRecords(
			data, "records", since, "updated_at", time.RFC3339, mockNextPageFunc,
		)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Empty(t, nextPage) // Last record processed is not the last in array
		assert.Equal(t, "1", result[0]["id"])
		assert.Equal(t, "2", result[1]["id"])
	}
}

func testAllRecordsOlder(createTestData func([]map[string]any) *ajson.Node,
	createRecord func(string, string) map[string]any, mockNextPageFunc func(*ajson.Node) (string, error),
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		records := []map[string]any{
			createRecord("1", "2022-12-31T10:00:00Z"),
			createRecord("2", "2022-12-30T10:00:00Z"),
			createRecord("3", "2022-12-29T10:00:00Z"),
		}
		data := createTestData(records)
		since := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		result, nextPage, err := FilterSortedRecords(
			data, "records", since, "updated_at", time.RFC3339, mockNextPageFunc,
		)

		assert.NoError(t, err)
		assert.Empty(t, result)
		assert.Empty(t, nextPage)
	}
}

func testLastRecordNewer(createTestData func([]map[string]any) *ajson.Node,
	createRecord func(string, string) map[string]any, mockNextPageFunc func(*ajson.Node) (string, error),
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		records := []map[string]any{
			createRecord("1", "2023-01-03T10:00:00Z"),
			createRecord("2", "2023-01-02T10:00:00Z"),
			createRecord("3", "2023-01-01T10:00:00Z"), // last record is newer
		}
		data := createTestData(records)
		since := time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC)

		result, nextPage, err := FilterSortedRecords(
			data, "records", since, "updated_at", time.RFC3339, mockNextPageFunc,
		)

		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, "next-page", nextPage)
	}
}

func testInvalidRecordsKey(createTestData func([]map[string]any) *ajson.Node,
	createRecord func(string, string) map[string]any, mockNextPageFunc func(*ajson.Node) (string, error),
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		data := createTestData([]map[string]any{createRecord("1", "2023-01-01T10:00:00Z")})
		since := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		_, _, err := FilterSortedRecords(
			data, "invalid_key", since, "updated_at", time.RFC3339, mockNextPageFunc,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bad records key")
	}
}

func testInvalidTimestampKey(createTestData func([]map[string]any) *ajson.Node,
	createRecord func(string, string) map[string]any, mockNextPageFunc func(*ajson.Node) (string, error),
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		data := createTestData([]map[string]any{createRecord("1", "2023-01-01T10:00:00Z")})
		since := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		_, _, err := FilterSortedRecords(
			data, "records", since, "invalid_timestamp_key", time.RFC3339, mockNextPageFunc,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bad since timestamp key")
	}
}

func testInvalidTimestampFormat(createTestData func([]map[string]any) *ajson.Node,
	mockNextPageFunc func(*ajson.Node) (string, error),
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		records := []map[string]any{{
			"id":         "1",
			"updated_at": "invalid-date-format",
			"name":       "record-1",
		}}
		data := createTestData(records)
		since := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		_, _, err := FilterSortedRecords(
			data, "records", since, "updated_at", time.RFC3339, mockNextPageFunc,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot parse timestamp")
	}
}

func testNextPageError(createTestData func([]map[string]any) *ajson.Node,
	createRecord func(string, string) map[string]any, errorNextPageFunc func(*ajson.Node) (string, error),
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		data := createTestData([]map[string]any{createRecord("1", "2023-01-01T10:00:00Z")})
		since := time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC)

		_, _, err := FilterSortedRecords(
			data, "records", since, "updated_at", time.RFC3339, errorNextPageFunc,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "constructing next page value")
	}
}

func testDifferentTimestampFormat(createTestData func([]map[string]any) *ajson.Node,
	mockNextPageFunc func(*ajson.Node) (string, error),
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		records := []map[string]any{{
			"id":         "1",
			"updated_at": "2023-01-01 10:00:00",
			"name":       "record-1",
		}}
		data := createTestData(records)
		since := time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC)

		result, nextPage, err := FilterSortedRecords(
			data, "records", since, "updated_at", "2006-01-02 15:04:05", mockNextPageFunc,
		)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "next-page", nextPage)
	}
}

func testExactSameTimestamp(createTestData func([]map[string]any) *ajson.Node,
	createRecord func(string, string) map[string]any, mockNextPageFunc func(*ajson.Node) (string, error),
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		records := []map[string]any{
			createRecord("1", "2023-01-01T10:00:00Z"), // exact same time
			createRecord("2", "2023-01-01T09:00:00Z"), // older
		}
		data := createTestData(records)
		since, _ := time.Parse(time.RFC3339, "2023-01-01T10:00:00Z")

		result, nextPage, err := FilterSortedRecords(
			data, "records", since, "updated_at", time.RFC3339, mockNextPageFunc,
		)

		assert.NoError(t, err)
		assert.Empty(t, result)
		assert.Empty(t, nextPage)
	}
}

func testComplexNestedJSON(mockNextPageFunc func(*ajson.Node) (string, error)) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		jsonStr := `{
			"metadata": {"count": 2, "page": 1},
			"records": [
				{"id": "1", "updated_at": "2023-01-02T10:00:00Z", "name": "record-1"},
				{"id": "2", "updated_at": "2023-01-01T10:00:00Z", "name": "record-2"}
			]
		}`

		node, err := ajson.Unmarshal([]byte(jsonStr))
		require.NoError(t, err)

		since := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		result, nextPage, err := FilterSortedRecords(
			node, "records", since, "updated_at", time.RFC3339, mockNextPageFunc,
		)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "next-page", nextPage)
	}
}
